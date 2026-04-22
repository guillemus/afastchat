package db

import (
	"context"
	"fmt"
	"time"

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
			WHERE id = :id
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":id": id},
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
			VALUES (:id, :title) 
			RETURNING id, title, archived_at, is_streaming, created_at, updated_at
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":id": id, ":title": title},
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
			SET title = :title, updated_at = datetime('now') 
			WHERE id = :id
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":title": title, ":id": id},
		})
	})
}

func (x *Pool) ArchiveThread(ctx context.Context, id string) error {
	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.ExecuteTransient(conn, `
			UPDATE threads 
			SET archived_at = datetime('now'), updated_at = datetime('now') 
			WHERE id = :id
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":id": id},
		})
	})
}

func (x *Pool) DeleteThread(ctx context.Context, id string) error {
	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.ExecuteTransient(conn, `
			DELETE FROM threads 
			WHERE id = :id
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":id": id},
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
			SET is_streaming = :is_streaming, updated_at = datetime('now') 
			WHERE id = :id
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":is_streaming": streamingInt, ":id": id},
		})
	})
}

func (x *Pool) CreateMessage(ctx context.Context, threadID string, typ MessageType, content string) (Message, error) {
	return x.CreateMessageWithMeta(ctx, Message{
		ThreadID: threadID,
		Type:     typ,
		Content:  content,
	})
}

func (x *Pool) CreateMessageWithMeta(ctx context.Context, message Message) (Message, error) {
	var created Message
	compressedContent := compressMessageContent(message.Content)
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		id := uuid.New().String()
		seq := 1

		err := sqlitex.Execute(conn, `
			SELECT COALESCE(MAX(seq) + 1, 1) AS next_seq
			FROM messages
			WHERE thread_id = :thread_id
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":thread_id": message.ThreadID},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				seq = int(stmt.GetInt64("next_seq"))
				return nil
			},
		})
		if err != nil {
			return fmt.Errorf("select next seq: %w", err)
		}

		err = sqlitex.Execute(conn, `
			INSERT INTO messages (id, thread_id, type, seq, content, tool_call_id, tool_name, provider_response_id) 
			VALUES (:id, :thread_id, :type, :seq, :content, :tool_call_id, :tool_name, :provider_response_id) 
			RETURNING id, thread_id, type, seq, content, tool_call_id, tool_name, provider_response_id, created_at, updated_at
		`, &sqlitex.ExecOptions{
			Named: map[string]any{
				":id":                   id,
				":thread_id":            message.ThreadID,
				":type":                 message.Type,
				":seq":                  seq,
				":content":              compressedContent,
				":tool_call_id":         message.ToolCallID,
				":tool_name":            message.ToolName,
				":provider_response_id": message.ProviderResponseID,
			},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				scannedMessage, err := scanMessage(stmt)
				if err != nil {
					return fmt.Errorf("scan message: %w", err)
				}
				created = scannedMessage
				return nil
			},
		})
		if err != nil {
			return fmt.Errorf("insert: %w", err)
		}
		return nil
	})

	return created, err
}

func (x *Pool) GetMessagesByThread(ctx context.Context, threadID string) ([]Message, error) {
	var messages []Message
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.Execute(conn, `
			SELECT id, thread_id, type, seq, content, tool_call_id, tool_name, provider_response_id, created_at, updated_at 
			FROM messages 
			WHERE thread_id = :thread_id 
			ORDER BY seq ASC
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":thread_id": threadID},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				message, err := scanMessage(stmt)
				if err != nil {
					return fmt.Errorf("scan message: %w", err)
				}
				messages = append(messages, message)
				return nil
			},
		})
	})
	return messages, err
}

func (x *Pool) UpdateMessageContent(ctx context.Context, id, content string) error {
	compressedContent := compressMessageContent(content)
	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.ExecuteTransient(conn, `
			UPDATE messages 
			SET content = :content, updated_at = datetime('now') 
			WHERE id = :id
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":content": compressedContent, ":id": id},
		})
	})
}

func (x *Pool) DeleteMessage(ctx context.Context, id string) error {
	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.ExecuteTransient(conn, `
			DELETE FROM messages 
			WHERE id = :id
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":id": id},
		})
	})
}

func (x *Pool) InitThread(ctx context.Context, title, sysMsg string) (string, error) {
	threadID := uuid.New().String()
	compressedSysMsg := compressMessageContent(sysMsg)
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		// Insert thread
		err := sqlitex.ExecuteTransient(conn, `
			INSERT INTO threads (id, title) 
			VALUES (:id, :title)
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":id": threadID, ":title": title},
		})
		if err != nil {
			return fmt.Errorf("create thread: %w", err)
		}

		// Insert system message
		err = sqlitex.ExecuteTransient(conn, `
			INSERT INTO messages (id, thread_id, type, seq, content) 
			VALUES (:id, :thread_id, 'system_message', 1, :content)
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":id": uuid.New().String(), ":thread_id": threadID, ":content": compressedSysMsg},
		})
		if err != nil {
			return fmt.Errorf("create system message: %w", err)
		}

		return nil
	})
	return threadID, err
}

func (x *Pool) CreateUserMessage(ctx context.Context, threadID, msg string) (Message, error) {
	return x.CreateMessage(ctx, threadID, "user_message", msg)
}

func (x *Pool) CreateAssistantMessage(ctx context.Context, threadID, msg string) (string, error) {
	m, err := x.CreateMessage(ctx, threadID, "assistant_message", msg)
	return m.ID, err
}

func (x *Pool) CreateSystemMessage(ctx context.Context, threadID, msg string) (string, error) {
	m, err := x.CreateMessage(ctx, threadID, "system_message", msg)
	return m.ID, err
}

func (x *Pool) GetUser(ctx context.Context, id string) (User, error) {
	var user User
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		err := sqlitex.Execute(conn, `
			SELECT id, email, password_hash, created_at, updated_at
			FROM users
			WHERE id = :id
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":id": id},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				user = scanUser(stmt)
				return nil
			},
		})
		if err != nil {
			return fmt.Errorf("query: %w", err)
		}
		if user.ID == "" {
			return ErrNotFound
		}
		return nil
	})
	return user, err
}

func (x *Pool) GetUserByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		err := sqlitex.Execute(conn, `
			SELECT id, email, password_hash, created_at, updated_at
			FROM users
			WHERE lower(email) = lower(:email)
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":email": email},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				user = scanUser(stmt)
				return nil
			},
		})
		if err != nil {
			return fmt.Errorf("query: %w", err)
		}
		if user.ID == "" {
			return ErrNotFound
		}
		return nil
	})
	return user, err
}

func (x *Pool) CreateUser(ctx context.Context, email, passwordHash string) (User, error) {
	id := uuid.New().String()
	var user User
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		err := sqlitex.Execute(conn, `
			INSERT INTO users (id, email, password_hash)
			VALUES (:id, :email, :password_hash)
			RETURNING id, email, password_hash, created_at, updated_at
		`, &sqlitex.ExecOptions{
			Named: map[string]any{
				":id":            id,
				":email":         email,
				":password_hash": passwordHash,
			},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				user = scanUser(stmt)
				return nil
			},
		})
		if err != nil {
			return fmt.Errorf("insert: %w", err)
		}
		return nil
	})
	return user, err
}

func (x *Pool) CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (Session, error) {
	id := uuid.New().String()
	var session Session
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		err := sqlitex.Execute(conn, `
			INSERT INTO sessions (id, user_id, token_hash, expires_at)
			VALUES (:id, :user_id, :token_hash, :expires_at)
			RETURNING id, user_id, token_hash, expires_at, created_at
		`, &sqlitex.ExecOptions{
			Named: map[string]any{
				":id":         id,
				":user_id":    userID,
				":token_hash": tokenHash,
				":expires_at": expiresAt.UTC().Format("2006-01-02 15:04:05"),
			},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				session = scanSession(stmt)
				return nil
			},
		})
		if err != nil {
			return fmt.Errorf("insert: %w", err)
		}
		return nil
	})
	return session, err
}

func (x *Pool) GetSessionByTokenHash(ctx context.Context, tokenHash string) (Session, error) {
	var session Session
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		err := sqlitex.Execute(conn, `
			SELECT id, user_id, token_hash, expires_at, created_at
			FROM sessions
			WHERE token_hash = :token_hash
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":token_hash": tokenHash},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				session = scanSession(stmt)
				return nil
			},
		})
		if err != nil {
			return fmt.Errorf("query: %w", err)
		}
		if session.ID == "" {
			return ErrNotFound
		}
		return nil
	})
	return session, err
}

func (x *Pool) DeleteSession(ctx context.Context, id string) error {
	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.ExecuteTransient(conn, `
			DELETE FROM sessions
			WHERE id = :id
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":id": id},
		})
	})
}

func (x *Pool) DeleteExpiredSessions(ctx context.Context, now time.Time) error {
	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.ExecuteTransient(conn, `
			DELETE FROM sessions
			WHERE expires_at <= :now
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":now": now.UTC().Format("2006-01-02 15:04:05")},
		})
	})
}
