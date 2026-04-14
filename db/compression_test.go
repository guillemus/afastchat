package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func TestCreateMessageCompressesAndReadsBack(t *testing.T) {
	t.Parallel()

	pool := newTestPool(t)
	ctx := context.Background()

	thread, err := pool.CreateThread(ctx, "thread")
	require.NoError(t, err)

	want := "hello hello hello hello hello"
	message, err := pool.CreateMessage(ctx, thread.ID, "user", want)
	require.NoError(t, err)
	require.Equal(t, want, message.Content)

	assertStoredContentIsCompressedBlob(t, pool, message.ID, want)

	messages, err := pool.GetMessagesByThread(ctx, thread.ID)
	require.NoError(t, err)
	require.Len(t, messages, 1)
	require.Equal(t, want, messages[0].Content)
}

func TestUpdateMessageContentRecompresses(t *testing.T) {
	t.Parallel()

	pool := newTestPool(t)
	ctx := context.Background()

	thread, err := pool.CreateThread(ctx, "thread")
	require.NoError(t, err)

	message, err := pool.CreateMessage(ctx, thread.ID, "assistant", "old")
	require.NoError(t, err)

	want := "new content new content new content"
	err = pool.UpdateMessageContent(ctx, message.ID, want)
	require.NoError(t, err)

	messages, err := pool.GetMessagesByThread(ctx, thread.ID)
	require.NoError(t, err)
	require.Len(t, messages, 1)
	require.Equal(t, want, messages[0].Content)

	assertStoredContentIsCompressedBlob(t, pool, message.ID, want)
}

func newTestPool(t *testing.T) *Pool {
	t.Helper()

	dbPath := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	pool, err := Connect(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, pool.Close())
	})

	err = pool.WithConn(context.Background(), func(conn *sqlite.Conn) error {
		return sqlitex.ExecuteScript(conn, schema, nil)
	})
	require.NoError(t, err)

	return pool
}

func assertStoredContentIsCompressedBlob(t *testing.T, pool *Pool, messageID, wantPlain string) {
	t.Helper()

	err := pool.WithConn(context.Background(), func(conn *sqlite.Conn) error {
		return sqlitex.Execute(conn, `
			SELECT content FROM messages WHERE id = ?
		`, &sqlitex.ExecOptions{
			Args: []any{messageID},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				require.Equal(t, sqlite.TypeBlob, stmt.ColumnType(stmt.ColumnIndex("content")))

				raw := make([]byte, stmt.GetLen("content"))
				stmt.GetBytes("content", raw)
				require.NotEqual(t, wantPlain, string(raw))
				require.Greater(t, len(raw), len(messageContentMagic))
				require.Equal(t, string(messageContentMagic), string(raw[:len(messageContentMagic)]))
				return nil
			},
		})
	})
	require.NoError(t, err)
}
