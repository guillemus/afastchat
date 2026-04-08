package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func (c *Conn) GetThread(ctx context.Context, id string) (Thread, error) {
	conn, err := c.pool.Take(ctx)
	if err != nil {
		return Thread{}, fmt.Errorf("take conn: %w", err)
	}
	defer c.pool.Put(conn)

	var thread Thread
	err = sqlitex.Execute(conn, `
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
		return Thread{}, fmt.Errorf("query: %w", err)
	}
	if thread.ID == "" {
		return Thread{}, ErrNotFound
	}
	return thread, nil
}

func (c *Conn) ListThreads(ctx context.Context) ([]Thread, error) {
	conn, err := c.pool.Take(ctx)
	if err != nil {
		return nil, fmt.Errorf("take conn: %w", err)
	}
	defer c.pool.Put(conn)

	var threads []Thread
	err = sqlitex.Execute(conn, `
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
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	return threads, nil
}

func (c *Conn) CreateThread(ctx context.Context, title string) (Thread, error) {
	conn, err := c.pool.Take(ctx)
	if err != nil {
		return Thread{}, fmt.Errorf("take conn: %w", err)
	}
	defer c.pool.Put(conn)

	id := uuid.New().String()
	var thread Thread
	err = sqlitex.Execute(conn, `
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
		return Thread{}, fmt.Errorf("insert: %w", err)
	}
	return thread, nil
}

func (c *Conn) UpdateThreadTitle(ctx context.Context, id, title string) error {
	conn, err := c.pool.Take(ctx)
	if err != nil {
		return fmt.Errorf("take conn: %w", err)
	}
	defer c.pool.Put(conn)

	return sqlitex.ExecuteTransient(conn, `
		UPDATE threads 
		SET title = ?, updated_at = datetime('now') 
		WHERE id = ?
	`, &sqlitex.ExecOptions{
		Args: []any{title, id},
	})
}

func (c *Conn) ArchiveThread(ctx context.Context, id string) error {
	conn, err := c.pool.Take(ctx)
	if err != nil {
		return fmt.Errorf("take conn: %w", err)
	}
	defer c.pool.Put(conn)

	return sqlitex.ExecuteTransient(conn, `
		UPDATE threads 
		SET archived_at = datetime('now'), updated_at = datetime('now') 
		WHERE id = ?
	`, &sqlitex.ExecOptions{
		Args: []any{id},
	})
}

func (c *Conn) DeleteThread(ctx context.Context, id string) error {
	conn, err := c.pool.Take(ctx)
	if err != nil {
		return fmt.Errorf("take conn: %w", err)
	}
	defer c.pool.Put(conn)

	return sqlitex.ExecuteTransient(conn, `
		DELETE FROM threads 
		WHERE id = ?
	`, &sqlitex.ExecOptions{
		Args: []any{id},
	})
}

func (c *Conn) SetThreadStreaming(ctx context.Context, id string, isStreaming bool) error {
	conn, err := c.pool.Take(ctx)
	if err != nil {
		return fmt.Errorf("take conn: %w", err)
	}
	defer c.pool.Put(conn)

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
}

func (c *Conn) CreateMessage(ctx context.Context, threadID, role, content string) (Message, error) {
	conn, err := c.pool.Take(ctx)
	if err != nil {
		return Message{}, fmt.Errorf("take conn: %w", err)
	}
	defer c.pool.Put(conn)

	id := uuid.New().String()
	var msg Message
	err = sqlitex.Execute(conn, `
		INSERT INTO messages (id, thread_id, role, content) 
		VALUES (?, ?, ?, ?) 
		RETURNING id, thread_id, role, content, created_at
	`, &sqlitex.ExecOptions{
		Args: []any{id, threadID, role, content},
		ResultFunc: func(stmt *sqlite.Stmt) error {
			msg = scanMessage(stmt)
			return nil
		},
	})
	if err != nil {
		return Message{}, fmt.Errorf("insert: %w", err)
	}
	return msg, nil
}

func (c *Conn) GetMessagesByThread(ctx context.Context, threadID string) ([]Message, error) {
	conn, err := c.pool.Take(ctx)
	if err != nil {
		return nil, fmt.Errorf("take conn: %w", err)
	}
	defer c.pool.Put(conn)

	var messages []Message
	err = sqlitex.Execute(conn, `
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
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	return messages, nil
}

func (c *Conn) UpdateMessageContent(ctx context.Context, id, content string) error {
	conn, err := c.pool.Take(ctx)
	if err != nil {
		return fmt.Errorf("take conn: %w", err)
	}
	defer c.pool.Put(conn)

	return sqlitex.ExecuteTransient(conn, `
		UPDATE messages 
		SET content = ? 
		WHERE id = ?
	`, &sqlitex.ExecOptions{
		Args: []any{content, id},
	})
}

func (c *Conn) DeleteMessage(ctx context.Context, id string) error {
	conn, err := c.pool.Take(ctx)
	if err != nil {
		return fmt.Errorf("take conn: %w", err)
	}
	defer c.pool.Put(conn)

	return sqlitex.ExecuteTransient(conn, `
		DELETE FROM messages 
		WHERE id = ?
	`, &sqlitex.ExecOptions{
		Args: []any{id},
	})
}

func (c *Conn) InitThread(ctx context.Context, title, sysMsg, userMsg string) (string, error) {
	conn, err := c.pool.Take(ctx)
	if err != nil {
		return "", fmt.Errorf("take conn: %w", err)
	}
	defer c.pool.Put(conn)

	threadID := uuid.New().String()

	// Insert thread
	err = sqlitex.ExecuteTransient(conn, `
		INSERT INTO threads (id, title) 
		VALUES (?, ?)
	`, &sqlitex.ExecOptions{
		Args: []any{threadID, title},
	})
	if err != nil {
		return "", fmt.Errorf("create thread: %w", err)
	}

	// Insert system message
	err = sqlitex.ExecuteTransient(conn, `
		INSERT INTO messages (id, thread_id, role, content) 
		VALUES (?, ?, ?, ?)
	`, &sqlitex.ExecOptions{
		Args: []any{uuid.New().String(), threadID, "system", sysMsg},
	})
	if err != nil {
		return "", fmt.Errorf("create system message: %w", err)
	}

	// Insert user message
	err = sqlitex.ExecuteTransient(conn, `
		INSERT INTO messages (id, thread_id, role, content) 
		VALUES (?, ?, ?, ?)
	`, &sqlitex.ExecOptions{
		Args: []any{uuid.New().String(), threadID, "user", userMsg},
	})
	if err != nil {
		return "", fmt.Errorf("create user message: %w", err)
	}

	return threadID, nil
}

func (c *Conn) CreateUserMessage(ctx context.Context, threadID, msg string) (Message, error) {
	return c.CreateMessage(ctx, threadID, "user", msg)
}

func (c *Conn) CreateAssistantMessage(ctx context.Context, threadID, msg string) (string, error) {
	m, err := c.CreateMessage(ctx, threadID, "assistant", msg)
	return m.ID, err
}

func (c *Conn) CreateSystemMessage(ctx context.Context, threadID, msg string) (string, error) {
	m, err := c.CreateMessage(ctx, threadID, "system", msg)
	return m.ID, err
}
