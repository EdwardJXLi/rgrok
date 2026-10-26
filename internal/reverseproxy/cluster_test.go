package reverseproxy

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxyForwarding(t *testing.T) {
	type forwardedRequest struct {
		Method         string
		Host           string
		RequestURI     string
		Body           string
		ForwardedFor   string
		ForwardedHost  string
		ForwardedProto string
		Forwarded      string
	}
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(forwardedRequest{
			Method: r.Method, Host: r.Host, RequestURI: r.RequestURI, Body: string(body),
			ForwardedFor:   r.Header.Get("X-Forwarded-For"),
			ForwardedHost:  r.Header.Get("X-Forwarded-Host"),
			ForwardedProto: r.Header.Get("X-Forwarded-Proto"),
			Forwarded:      r.Header.Get("Forwarded"),
		})
	}))
	t.Cleanup(backend.Close)
	cluster := NewCluster()
	_, err := cluster.Set("staging.example.com", "fallback.example.com", backend.Listener.Addr().String())
	require.NoError(t, err)
	proxy, ok := cluster.Get("staging.example.com")
	require.True(t, ok)

	tests := []struct {
		name    string
		scheme  string
		query   string
		ingress bool
	}{
		{name: "HTTP", scheme: "http", query: "key=one&key=two"},
		{name: "HTTPS", scheme: "https", query: "key=one&key=two"},
		{name: "ingress proxy", scheme: "http", query: "key=one&key=two", ingress: true},
		{name: "unparsed query", scheme: "http", query: "filter=one;two&key=three"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requestURI := "/api/a%2Fb?" + test.query
			req := httptest.NewRequest(http.MethodPost, test.scheme+"://staging.example.com"+requestURI, strings.NewReader("request body"))
			req.RemoteAddr = "192.0.2.1:1234"
			want := forwardedRequest{
				Method: http.MethodPost, Host: "staging.example.com", RequestURI: requestURI, Body: "request body",
				ForwardedFor: "192.0.2.1", ForwardedHost: "staging.example.com", ForwardedProto: test.scheme,
			}
			if test.ingress {
				req.Header.Set("X-Forwarded-For", "198.51.100.1")
				req.Header.Set("X-Forwarded-Host", "public.example.com")
				req.Header.Set("X-Forwarded-Proto", "https")
				req.Header.Set("Forwarded", "for=198.51.100.1;proto=https")
				want.ForwardedFor = "198.51.100.1, 192.0.2.1"
				want.ForwardedHost = "public.example.com"
				want.ForwardedProto = "https"
				want.Forwarded = "for=198.51.100.1;proto=https"
			}
			response := httptest.NewRecorder()
			proxy.ServeHTTP(response, req)
			require.Equal(t, http.StatusCreated, response.Code)
			assert.Equal(t, "application/json", response.Header().Get("Content-Type"))
			var got forwardedRequest
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &got))
			assert.Equal(t, want, got)
		})
	}
}
