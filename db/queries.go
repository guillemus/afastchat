package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"zombiezen.com/go/sqlite/sqlitex"
)

func (c *Conn) GetThread(ctx context.Context, id string) (Thread, error) {
	conn, err := c.pool.Take(ctx)
	if err != nil {
		return Thread{}, fmt.Errorf("take conn: %w", err)
	}
	defer c.pool.Put(conn)

	stmt, err := conn.Prepare(`
		SELECT id, title, archived_at, is_streaming, created_at, updated_at 
		FROM threads 
		WHERE id = ?
	`)
	if err != nil {
		return Thread{}, fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Finalize()

	stmt.BindText(1, id)
	hasRow, err := stmt.Step()
	if err != nil {
		return Thread{}, fmt.Errorf("step: %w", err)
	}
	if !hasRow {
		return Thread{}, ErrNotFound
	}

	return scanThread(stmt), nil
}

func (c *Conn) ListThreads(ctx context.Context) ([]Thread, error) {
	conn, err := c.pool.Take(ctx)
	if err != nil {
		return nil, fmt.Errorf("take conn: %w", err)
	}
	defer c.pool.Put(conn)

	stmt, err := conn.Prepare(`
		SELECT id, title, archived_at, is_streaming, created_at, updated_at 
		FROM threads 
		WHERE archived_at IS NULL 
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Finalize()

	var threads []Thread
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return nil, fmt.Errorf("step: %w", err)
		}
		if !hasRow {
			break
		}
		threads = append(threads, scanThread(stmt))
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
	stmt, err := conn.Prepare(`
		INSERT INTO threads (id, title) 
		VALUES (?, ?) 
		RETURNING id, title, archived_at, is_streaming, created_at, updated_at
	`)
	if err != nil {
		return Thread{}, fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Finalize()

	stmt.BindText(1, id)
	stmt.BindText(2, title)

	hasRow, err := stmt.Step()
	if err != nil {
		return Thread{}, fmt.Errorf("insert: %w", err)
	}
	if !hasRow {
		return Thread{}, fmt.Errorf("no row returned")
	}

	return scanThread(stmt), nil
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
	stmt, err := conn.Prepare(`
		INSERT INTO messages (id, thread_id, role, content) 
		VALUES (?, ?, ?, ?) 
		RETURNING id, thread_id, role, content, created_at
	`)
	if err != nil {
		return Message{}, fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Finalize()

	stmt.BindText(1, id)
	stmt.BindText(2, threadID)
	stmt.BindText(3, role)
	stmt.BindText(4, content)

	hasRow, err := stmt.Step()
	if err != nil {
		return Message{}, fmt.Errorf("insert: %w", err)
	}
	if !hasRow {
		return Message{}, fmt.Errorf("no row returned")
	}

	return scanMessage(stmt), nil
}

func (c *Conn) GetMessagesByThread(ctx context.Context, threadID string) ([]Message, error) {
	conn, err := c.pool.Take(ctx)
	if err != nil {
		return nil, fmt.Errorf("take conn: %w", err)
	}
	defer c.pool.Put(conn)

	stmt, err := conn.Prepare(`
		SELECT id, thread_id, role, content, created_at 
		FROM messages 
		WHERE thread_id = ? 
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Finalize()

	stmt.BindText(1, threadID)

	var messages []Message
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return nil, fmt.Errorf("step: %w", err)
		}
		if !hasRow {
			break
		}
		messages = append(messages, scanMessage(stmt))
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

	err = sqlitex.Execute(conn, `
		INSERT INTO threads (id, title) 
		VALUES (?, ?)
	`, &sqlitex.ExecOptions{
		Args: []any{threadID, title},
	})
	if err != nil {
		return "", fmt.Errorf("create thread: %w", err)
	}

	err = sqlitex.Execute(conn, `
		INSERT INTO messages (id, thread_id, role, content) 
		VALUES (?, ?, ?, ?)
	`, &sqlitex.ExecOptions{
		Args: []any{uuid.New().String(), threadID, "system", sysMsg},
	})
	if err != nil {
		return "", fmt.Errorf("create system message: %w", err)
	}

	err = sqlitex.Execute(conn, `
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
