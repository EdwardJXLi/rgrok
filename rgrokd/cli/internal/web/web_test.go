package web

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/flamego/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"unknwon.dev/x/logx"

	"github.com/EdwardJXLi/rgrok/internal/conf"
	"github.com/EdwardJXLi/rgrok/internal/database"
)

func TestExistingSQLiteSession(t *testing.T) {
	ctx := context.Background()
	logger := logx.New(slog.NewTextHandler(io.Discard, nil))
	dir := t.TempDir()
	config := &conf.Config{
		Database: &conf.Database{Type: conf.DatabaseTypeSQLite, Path: filepath.Join(dir, "rgrok.db")},
		Proxy:    conf.Proxy{Scheme: "https", Domain: "example.com"},
	}
	db, err := database.New(logger, config.Database)
	require.NoError(t, err)
	sqlDB, err := db.DB.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	principal, err := db.UpsertPrincipal(ctx, database.UpsertPrincipalOptions{
		Identifier: "alice", DisplayName: "Alice", Token: "test-token", Subdomain: "alice",
	})
	require.NoError(t, err)

	store, err := session.FileIniter()(ctx,
		session.FileConfig{RootDir: filepath.Join(dir, "rgrok-sessions")},
		session.IDWriter(func(http.ResponseWriter, *http.Request, string) {}),
	)
	require.NoError(t, err)
	const sessionID = "existing12345678"
	sess, err := store.Read(ctx, sessionID)
	require.NoError(t, err)
	sess.Set("userID", principal.ID)
	require.NoError(t, store.Save(ctx, sess))

	server, err := NewServer(logger, config, db)
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodGet, "/api/user-info", nil)
	request.AddCookie(&http.Cookie{Name: "rgrokd_session", Value: sessionID})
	response := httptest.NewRecorder()
	server.Handler.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	var userInfo map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &userInfo))
	assert.Equal(t, map[string]string{
		"displayName": "Alice", "token": "test-token", "url": "https://alice.example.com",
	}, userInfo)
}
