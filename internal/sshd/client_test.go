package sshd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	charmlog "charm.land/log/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"unknwon.dev/x/logx"

	"github.com/EdwardJXLi/rgrok/internal/conf"
	"github.com/EdwardJXLi/rgrok/internal/database"
	"github.com/EdwardJXLi/rgrok/internal/reverseproxy"
)

type testServer struct {
	addr    string
	proxies *reverseproxy.Cluster
}

func startTestServer(t *testing.T, strategy string) *testServer {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "rgrokd.yml")
	require.NoError(t, os.WriteFile(configPath, []byte(fmt.Sprintf("proxy:\n  domain: tunnels.example.test:3000\n  scheme: http\n  subdomain_strategy: %q\n", strategy)), 0600))
	config, err := conf.Load(configPath)
	require.NoError(t, err)
	logger := logx.New(charmlog.New(io.Discard))
	db, err := database.New(logger, &conf.Database{Type: conf.DatabaseTypeSQLite, Path: filepath.Join(t.TempDir(), "test.db")})
	require.NoError(t, err)
	sqlDB, err := db.DB.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	_, err = db.UpsertPrincipal(t.Context(), database.UpsertPrincipalOptions{
		Identifier: "alice", DisplayName: "Alice", Token: "test-token", Subdomain: "alice",
	})
	require.NoError(t, err)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := listener.Addr().String()
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	server := &testServer{addr: addr, proxies: reverseproxy.NewCluster()}
	go func() { done <- Start(ctx, logger, port, config.Proxy, db, server.proxies) }()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Error("SSH server did not stop")
		}
	})
	return server
}

func (s *testServer) dial(t *testing.T) *ssh.Client {
	t.Helper()
	var client *ssh.Client
	var err error
	require.Eventually(t, func() bool {
		client, err = ssh.Dial("tcp", s.addr, &ssh.ClientConfig{
			User: "pgrok", Auth: []ssh.AuthMethod{ssh.Password("test-token")},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), Timeout: time.Second,
		})
		return err == nil
	}, 5*time.Second, 10*time.Millisecond)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func sendHint(t *testing.T, client *ssh.Client, protocol, subdomain string) (bool, string) {
	t.Helper()
	payload, err := json.Marshal(map[string]string{"protocol": protocol, "subdomain": subdomain})
	require.NoError(t, err)
	ok, reply, err := client.SendRequest("hint", true, payload)
	require.NoError(t, err)
	return ok, string(reply)
}

func openHTTPTunnel(t *testing.T, client *ssh.Client) (net.Listener, string) {
	t.Helper()
	listener, err := client.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })
	ok, reply, err := client.SendRequest("server-info", true, nil)
	require.NoError(t, err)
	require.True(t, ok, string(reply))
	var info struct {
		HostURL string `json:"host_url"`
	}
	require.NoError(t, json.Unmarshal(reply, &info))
	return listener, strings.TrimPrefix(info.HostURL, "http://")
}

func (s *testServer) assertForward(t *testing.T, listener net.Listener, host string) {
	t.Helper()
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "forwarded "+host)
	})}
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { _ = server.Close() })
	proxy, ok := s.proxies.Get(host)
	require.True(t, ok)
	response := httptest.NewRecorder()
	proxy.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "http://"+host+"/", nil))
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "forwarded "+host, response.Body.String())
}

func TestHTTPSubdomainSelection(t *testing.T) {
	for _, strategy := range []string{"", "random", "stable"} {
		t.Run("strategy="+strategy, func(t *testing.T) {
			server := startTestServer(t, strategy)
			first := server.dial(t)
			// Vanilla SSH clients do not send a hint.
			_, firstHost := openHTTPTunnel(t, first)
			if strategy == "stable" {
				assert.Equal(t, "alice.tunnels.example.test:3000", firstHost)
			} else {
				assert.Regexp(t, `^[a-z][a-z0-9]{7}\.tunnels\.example\.test:3000$`, firstHost)
			}
			second := server.dial(t)
			ok, reply := sendHint(t, second, "http", "")
			require.True(t, ok, reply)
			_, secondHost := openHTTPTunnel(t, second)
			assert.NotEqual(t, firstHost, secondHost)
			if strategy == "stable" {
				assert.Regexp(t, `^[a-f0-9]{32}-alice\.tunnels\.example\.test:3000$`, secondHost)
			} else {
				assert.Regexp(t, `^[a-z][a-z0-9]{7}\.tunnels\.example\.test:3000$`, secondHost)
			}
			custom := server.dial(t)
			ok, reply = sendHint(t, custom, "http", "staging")
			require.True(t, ok, reply)
			listener, customHost := openHTTPTunnel(t, custom)
			assert.Equal(t, "staging.tunnels.example.test:3000", customHost)
			server.assertForward(t, listener, customHost)
		})
	}
}

func TestCustomSubdomainValidation(t *testing.T) {
	server := startTestServer(t, "random")
	for _, name := range []string{"Uppercase", "a.b", "-a", "a-", "a_b", "a/b", "a:80", "a\n", "é", strings.Repeat("a", 64)} {
		t.Run(name, func(t *testing.T) {
			client := server.dial(t)
			ok, reply := sendHint(t, client, "http", name)
			require.False(t, ok)
			assert.Contains(t, reply, "invalid subdomain")
		})
	}
	t.Run("TCP rejects custom names", func(t *testing.T) {
		ok, reply := sendHint(t, server.dial(t), "tcp", "staging")
		require.False(t, ok)
		assert.Contains(t, reply, "only supported for http")
	})
	for _, name := range []string{"0", "a-b", strings.Repeat("a", 63)} {
		t.Run(name, func(t *testing.T) {
			client := server.dial(t)
			ok, reply := sendHint(t, client, "http", name)
			require.True(t, ok, reply)
			_, host := openHTTPTunnel(t, client)
			assert.Equal(t, name+".tunnels.example.test:3000", host)
		})
	}
}

func TestCustomSubdomainLifecycle(t *testing.T) {
	server := startTestServer(t, "random")
	first := server.dial(t)
	ok, reply := sendHint(t, first, "http", "staging")
	require.True(t, ok, reply)
	listener, host := openHTTPTunnel(t, first)
	second := server.dial(t)
	ok, reply = sendHint(t, second, "http", "staging")
	require.True(t, ok, reply)
	_, err := second.Listen("tcp", "127.0.0.1:0")
	require.Error(t, err)
	ok, payload, err := second.SendRequest("server-info", true, nil)
	require.NoError(t, err)
	require.False(t, ok)
	assert.Contains(t, string(payload), "already in use")

	ok, reply = sendHint(t, first, "http", "other")
	require.False(t, ok)
	assert.Contains(t, reply, "already forwarding")
	server.assertForward(t, listener, host)
	require.NoError(t, listener.Close())
	require.Eventually(t, func() bool {
		_, exists := server.proxies.Get(host)
		return !exists
	}, time.Second, 10*time.Millisecond)

	third := server.dial(t)
	ok, reply = sendHint(t, third, "http", "staging")
	require.True(t, ok, reply)
	_, reusedHost := openHTTPTunnel(t, third)
	assert.Equal(t, host, reusedHost)
}
