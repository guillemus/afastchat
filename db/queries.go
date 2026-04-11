package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func (x *Pool) GetThread(ctx context.Context, id string) (Thread, error) {
	var thread Thread
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		err := sqlitex.Execute(conn, `
			SELECT id, title, archived_at, is_streaming, created_at, updated_at 
			FROM threads 
			WHERE id = ?
		`, &sqlitex.ExecOptions{
			Args: []any{id},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				thread = scanThread(stmt)
				return nil
			},
		})
		if err != nil {
			return fmt.Errorf("query: %w", err)
		}
		if thread.ID == "" {
			return ErrNotFound
		}
		return nil
	})
	return thread, err
}

func (x *Pool) ListThreads(ctx context.Context) ([]Thread, error) {
	var threads []Thread
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.Execute(conn, `
			SELECT id, title, archived_at, is_streaming, created_at, updated_at 
			FROM threads 
			WHERE archived_at IS NULL 
			ORDER BY created_at DESC
		`, &sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				threads = append(threads, scanThread(stmt))
				return nil
			},
		})
	})
	return threads, err
}

func (x *Pool) CreateThread(ctx context.Context, title string) (Thread, error) {
	id := uuid.New().String()
	var thread Thread
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		err := sqlitex.Execute(conn, `
			INSERT INTO threads (id, title) 
			VALUES (?, ?) 
			RETURNING id, title, archived_at, is_streaming, created_at, updated_at
		`, &sqlitex.ExecOptions{
			Args: []any{id, title},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				thread = scanThread(stmt)
				return nil
			},
		})
		if err != nil {
			return fmt.Errorf("insert: %w", err)
		}
		return nil
	})
	return thread, err
}

func (x *Pool) UpdateThreadTitle(ctx context.Context, id, title string) error {
	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.ExecuteTransient(conn, `
			UPDATE threads 
			SET title = ?, updated_at = datetime('now') 
			WHERE id = ?
		`, &sqlitex.ExecOptions{
			Args: []any{title, id},
		})
	})
}

func (x *Pool) ArchiveThread(ctx context.Context, id string) error {
	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.ExecuteTransient(conn, `
			UPDATE threads 
			SET archived_at = datetime('now'), updated_at = datetime('now') 
			WHERE id = ?
		`, &sqlitex.ExecOptions{
			Args: []any{id},
		})
	})
}

func (x *Pool) DeleteThread(ctx context.Context, id string) error {
	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.ExecuteTransient(conn, `
			DELETE FROM threads 
			WHERE id = ?
		`, &sqlitex.ExecOptions{
			Args: []any{id},
		})
	})
}

func (x *Pool) SetThreadStreaming(ctx context.Context, id string, isStreaming bool) error {
	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		streamingInt := 0
		if isStreaming {
			streamingInt = 1
		}

		return sqlitex.ExecuteTransient(conn, `
			UPDATE threads 
			SET is_streaming = ?, updated_at = datetime('now') 
			WHERE id = ?
		`, &sqlitex.ExecOptions{
			Args: []any{streamingInt, id},
		})
	})
}

func (x *Pool) CreateMessage(ctx context.Context, threadID, role, content string) (Message, error) {
	var message Message
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		id := uuid.New().String()
		err := sqlitex.Execute(conn, `
		INSERT INTO messages (id, thread_id, role, content) 
		VALUES (?, ?, ?, ?) 
		RETURNING id, thread_id, role, content, created_at
	`, &sqlitex.ExecOptions{
			Args: []any{id, threadID, role, content},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				message = scanMessage(stmt)
				return nil
			},
		})
		if err != nil {
			return fmt.Errorf("insert: %w", err)
		}
		return nil
	})

	return message, err
}

func (x *Pool) GetMessagesByThread(ctx context.Context, threadID string) ([]Message, error) {
	var messages []Message
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.Execute(conn, `
			SELECT id, thread_id, role, content, created_at 
			FROM messages 
			WHERE thread_id = ? 
			ORDER BY created_at ASC
		`, &sqlitex.ExecOptions{
			Args: []any{threadID},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				messages = append(messages, scanMessage(stmt))
				return nil
			},
		})
	})
	return messages, err
}

func (x *Pool) UpdateMessageContent(ctx context.Context, id, content string) error {
	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.ExecuteTransient(conn, `
			UPDATE messages 
			SET content = ? 
			WHERE id = ?
		`, &sqlitex.ExecOptions{
			Args: []any{content, id},
		})
	})
}

func (x *Pool) DeleteMessage(ctx context.Context, id string) error {
	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.ExecuteTransient(conn, `
			DELETE FROM messages 
			WHERE id = ?
		`, &sqlitex.ExecOptions{
			Args: []any{id},
		})
	})
}

func (x *Pool) InitThread(ctx context.Context, title, sysMsg string) (string, error) {
	threadID := uuid.New().String()
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		// Insert thread
		err := sqlitex.ExecuteTransient(conn, `
			INSERT INTO threads (id, title) 
			VALUES (?, ?)
		`, &sqlitex.ExecOptions{
			Args: []any{threadID, title},
		})
		if err != nil {
			return fmt.Errorf("create thread: %w", err)
		}

		// Insert system message
		err = sqlitex.ExecuteTransient(conn, `
			INSERT INTO messages (id, thread_id, role, content) 
			VALUES (?, ?, ?, ?)
		`, &sqlitex.ExecOptions{
			Args: []any{uuid.New().String(), threadID, "system", sysMsg},
		})
		if err != nil {
			return fmt.Errorf("create system message: %w", err)
		}

		return nil
	})
	return threadID, err
}

func (x *Pool) CreateUserMessage(ctx context.Context, threadID, msg string) (Message, error) {
	return x.CreateMessage(ctx, threadID, "user", msg)
}

func (x *Pool) CreateAssistantMessage(ctx context.Context, threadID, msg string) (string, error) {
	m, err := x.CreateMessage(ctx, threadID, "assistant", msg)
	return m.ID, err
}

func (x *Pool) CreateSystemMessage(ctx context.Context, threadID, msg string) (string, error) {
	m, err := x.CreateMessage(ctx, threadID, "system", msg)
	return m.ID, err
}
