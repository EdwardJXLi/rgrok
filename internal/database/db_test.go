package database

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"testing"

	charmlog "charm.land/log/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"unknwon.dev/x/logx"

	"github.com/EdwardJXLi/rgrok/internal/conf"
	"github.com/EdwardJXLi/rgrok/internal/cryptoutil"
)

func TestSQLitePersistence(t *testing.T) {
	logger := logx.New(charmlog.New(io.Discard))
	config := &conf.Database{Type: conf.DatabaseTypeSQLite, Path: filepath.Join(t.TempDir(), "rgrokd.db")}
	db, err := New(logger, config)
	require.NoError(t, err)
	sqlDB, err := db.DB.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	ctx := context.Background()
	principal, err := db.UpsertPrincipal(ctx, UpsertPrincipalOptions{
		Identifier: "alice", DisplayName: "Alice", Token: "token", Subdomain: "alice",
	})
	require.NoError(t, err)
	err = db.UpdatePrincipalLastTCPPort(ctx, principal.ID, 12345)
	require.NoError(t, err)
	key, err := db.CreateHostKey(ctx, cryptoutil.KeyAlgorithmEd25519, []byte("host-key"))
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	db, err = New(logger, config)
	require.NoError(t, err)
	reopened, err := db.DB.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = reopened.Close() })

	storedPrincipal, err := db.GetPrincipalByToken(ctx, "token")
	require.NoError(t, err)
	assert.Equal(t, principal.ID, storedPrincipal.ID)
	assert.Equal(t, "alice", storedPrincipal.Subdomain)
	assert.Equal(t, "Alice", storedPrincipal.DisplayName)
	assert.Equal(t, 12345, storedPrincipal.LastTCPPort)
	storedKey, err := db.GetHostKeyByAlgorithm(ctx, key.Algorithm)
	require.NoError(t, err)
	assert.Equal(t, key.PEM, storedKey.PEM)
}

func TestGORMLogger(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   string
	}{
		{
			name:   "normal",
			format: "something",
			want:   "DEBU something\n",
		},
		{
			name:   "error",
			format: "[error] oops",
			want:   "ERRO [error] oops\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer

			logger := &gormLogger{
				Logger: logx.New(charmlog.NewWithOptions(&buf, charmlog.Options{Level: charmlog.DebugLevel})),
			}
			logger.Printf("%s", test.format)
			assert.Equal(t, test.want, buf.String())
		})
	}
}
