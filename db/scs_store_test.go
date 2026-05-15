package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func newTestPool(t *testing.T) *Pool {
	t.Helper()

	pool, err := Connect("file:" + uuid.New().String() + "?mode=memory&cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })

	err = pool.WithConn(context.Background(), func(conn *sqlite.Conn) error {
		return sqlitex.ExecuteScript(conn, schema, nil)
	})
	require.NoError(t, err)

	return pool
}

func TestSCSStoreCommitFindDelete(t *testing.T) {
	t.Parallel()

	pool := newTestPool(t)
	store := NewSCSStoreWithCleanupInterval(pool, 0)

	err := store.CommitCtx(context.Background(), "token-1", []byte("hello"), time.Now().Add(time.Hour))
	require.NoError(t, err)

	data, found, err := store.FindCtx(context.Background(), "token-1")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, []byte("hello"), data)

	err = store.DeleteCtx(context.Background(), "token-1")
	require.NoError(t, err)

	data, found, err = store.FindCtx(context.Background(), "token-1")
	require.NoError(t, err)
	require.False(t, found)
	require.Nil(t, data)
}

func TestSCSStoreSkipsExpiredSessions(t *testing.T) {
	t.Parallel()

	pool := newTestPool(t)
	store := NewSCSStoreWithCleanupInterval(pool, 0)

	err := store.Commit("expired", []byte("old"), time.Now().Add(-time.Hour))
	require.NoError(t, err)

	data, found, err := store.Find("expired")
	require.NoError(t, err)
	require.False(t, found)
	require.Nil(t, data)

	sessions, err := store.All()
	require.NoError(t, err)
	require.Empty(t, sessions)
}

func TestSCSStoreAllReturnsActiveSessions(t *testing.T) {
	t.Parallel()

	pool := newTestPool(t)
	store := NewSCSStoreWithCleanupInterval(pool, 0)

	require.NoError(t, store.Commit("active-1", []byte("one"), time.Now().Add(time.Hour)))
	require.NoError(t, store.Commit("active-2", []byte("two"), time.Now().Add(time.Hour)))
	require.NoError(t, store.Commit("expired", []byte("old"), time.Now().Add(-time.Hour)))

	sessions, err := store.All()
	require.NoError(t, err)
	require.Equal(t, map[string][]byte{
		"active-1": []byte("one"),
		"active-2": []byte("two"),
	}, sessions)
}
