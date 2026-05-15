package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func (x *Pool) GetThread(ctx context.Context, id, userID string) (Thread, error) {
	return queryOne(ctx, x, "query thread", `
		SELECT id, user_id, title, archived_at, is_streaming, created_at, updated_at
		FROM threads
		WHERE id = :id AND user_id = :user_id
	`, map[string]any{":id": id, ":user_id": userID}, func(stmt *sqlite.Stmt) (Thread, error) {
		return scanThread(stmt), nil
	})
}

func (x *Pool) ListThreads(ctx context.Context, userID string) ([]Thread, error) {
	return queryAll(ctx, x, "query threads", `
		SELECT id, user_id, title, archived_at, is_streaming, created_at, updated_at
		FROM threads
		WHERE user_id = :user_id AND archived_at IS NULL
		ORDER BY created_at DESC
	`, map[string]any{":user_id": userID}, func(stmt *sqlite.Stmt) (Thread, error) {
		return scanThread(stmt), nil
	})
}

func (x *Pool) CreateThread(ctx context.Context, userID, title string) (Thread, error) {
	id := uuid.New().String()
	return execReturning(ctx, x, "insert thread", `
		INSERT INTO threads (id, user_id, title)
		VALUES (:id, :user_id, :title)
		RETURNING id, user_id, title, archived_at, is_streaming, created_at, updated_at
	`, map[string]any{":id": id, ":user_id": userID, ":title": title}, func(stmt *sqlite.Stmt) (Thread, error) {
		return scanThread(stmt), nil
	})
}

func (x *Pool) UpdateThreadTitle(ctx context.Context, userID, id, title string) error {
	return x.exec(ctx, "update thread title", `
		UPDATE threads
		SET title = :title, updated_at = datetime('now')
		WHERE id = :id AND user_id = :user_id
		RETURNING id
	`, map[string]any{":title": title, ":id": id, ":user_id": userID})
}

func (x *Pool) ArchiveThread(ctx context.Context, userID, id string) error {
	return x.exec(ctx, "archive thread", `
		UPDATE threads
		SET archived_at = datetime('now'), updated_at = datetime('now')
		WHERE id = :id AND user_id = :user_id
		RETURNING id
	`, map[string]any{":id": id, ":user_id": userID})
}

func (x *Pool) DeleteThread(ctx context.Context, userID, id string) error {
	return x.exec(ctx, "delete thread", `
		DELETE FROM threads
		WHERE id = :id AND user_id = :user_id
		RETURNING id
	`, map[string]any{":id": id, ":user_id": userID})
}

func (x *Pool) SetThreadStreaming(ctx context.Context, userID, id string, isStreaming bool) error {
	streamingInt := 0
	if isStreaming {
		streamingInt = 1
	}

	return x.exec(ctx, "set thread streaming", `
		UPDATE threads
		SET is_streaming = :is_streaming, updated_at = datetime('now')
		WHERE id = :id AND user_id = :user_id
		RETURNING id
	`, map[string]any{":is_streaming": streamingInt, ":id": id, ":user_id": userID})
}

func (x *Pool) CreateMessage(ctx context.Context, userID, threadID string, typ MessageType, content string) (Message, error) {
	return x.CreateMessageWithMeta(ctx, userID, Message{
		ThreadID: threadID,
		Type:     typ,
		Content:  content,
	})
}

func (x *Pool) CreateMessageWithMeta(ctx context.Context, userID string, message Message) (Message, error) {
	var created Message
	compressedContent := compressMessageContent(message.Content)
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		id := uuid.New().String()
		seq := 1

		err := sqlitex.Execute(conn, `
			SELECT COALESCE(MAX(messages.seq) + 1, 1) AS next_seq
			FROM messages
			JOIN threads ON threads.id = messages.thread_id
			WHERE messages.thread_id = :thread_id AND threads.user_id = :user_id
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":thread_id": message.ThreadID, ":user_id": userID},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				seq = int(stmt.GetInt64("next_seq"))
				return nil
			},
		})
		if err != nil {
			return fmt.Errorf("select next seq: %w", err)
		}

		threadExists := false
		err = sqlitex.Execute(conn, `
			SELECT id
			FROM threads
			WHERE id = :thread_id AND user_id = :user_id
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":thread_id": message.ThreadID, ":user_id": userID},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				threadExists = stmt.GetText("id") != ""
				return nil
			},
		})
		if err != nil {
			return fmt.Errorf("query thread: %w", err)
		}
		if !threadExists {
			return ErrNotFound
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

func (x *Pool) GetMessagesByThread(ctx context.Context, userID, threadID string) ([]Message, error) {
	return queryAll(ctx, x, "query messages", `
		SELECT messages.id, messages.thread_id, messages.type, messages.seq, messages.content,
		       messages.tool_call_id, messages.tool_name, messages.provider_response_id,
		       messages.created_at, messages.updated_at
		FROM messages
		JOIN threads ON threads.id = messages.thread_id
		WHERE messages.thread_id = :thread_id AND threads.user_id = :user_id
		ORDER BY messages.seq ASC
	`, map[string]any{":thread_id": threadID, ":user_id": userID}, func(stmt *sqlite.Stmt) (Message, error) {
		message, err := scanMessage(stmt)
		if err != nil {
			return Message{}, fmt.Errorf("scan message: %w", err)
		}
		return message, nil
	})
}

func (x *Pool) UpdateMessageContent(ctx context.Context, id, content string) error {
	compressedContent := compressMessageContent(content)
	return x.exec(ctx, "update message", `
		UPDATE messages
		SET content = :content, updated_at = datetime('now')
		WHERE id = :id
	`, map[string]any{":content": compressedContent, ":id": id})
}

func (x *Pool) DeleteMessage(ctx context.Context, id string) error {
	return x.exec(ctx, "delete message", `
		DELETE FROM messages
		WHERE id = :id
	`, map[string]any{":id": id})
}

func (x *Pool) InitThread(ctx context.Context, userID, title, sysMsg string) (string, error) {
	threadID := uuid.New().String()
	compressedSysMsg := compressMessageContent(sysMsg)
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		err := sqlitex.ExecuteTransient(conn, `
			INSERT INTO threads (id, user_id, title)
			VALUES (:id, :user_id, :title)
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":id": threadID, ":user_id": userID, ":title": title},
		})
		if err != nil {
			return fmt.Errorf("create thread: %w", err)
		}

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

func (x *Pool) CreateUserMessage(ctx context.Context, userID, threadID, msg string) (Message, error) {
	return x.CreateMessage(ctx, userID, threadID, "user_message", msg)
}

func (x *Pool) CreateAssistantMessage(ctx context.Context, userID, threadID, msg string) (string, error) {
	m, err := x.CreateMessage(ctx, userID, threadID, "assistant_message", msg)
	return m.ID, err
}

func (x *Pool) CreateSystemMessage(ctx context.Context, userID, threadID, msg string) (string, error) {
	m, err := x.CreateMessage(ctx, userID, threadID, "system_message", msg)
	return m.ID, err
}

func (x *Pool) GetUser(ctx context.Context, id string) (User, error) {
	return queryOne(ctx, x, "query user", `
		SELECT id, email, password_hash, paid_at, payment_customer_id, created_at, updated_at
		FROM users
		WHERE id = :id
	`, map[string]any{":id": id}, func(stmt *sqlite.Stmt) (User, error) {
		return scanUser(stmt), nil
	})
}

func (x *Pool) GetUserByEmail(ctx context.Context, email string) (User, error) {
	return queryOne(ctx, x, "query user by email", `
		SELECT id, email, password_hash, paid_at, payment_customer_id, created_at, updated_at
		FROM users
		WHERE lower(email) = lower(:email)
	`, map[string]any{":email": email}, func(stmt *sqlite.Stmt) (User, error) {
		return scanUser(stmt), nil
	})
}

func (x *Pool) CreateUser(ctx context.Context, email, passwordHash string) (User, error) {
	id := uuid.New().String()
	user, err := execReturning(ctx, x, "insert user", `
		INSERT INTO users (id, email, password_hash)
		VALUES (:id, :email, :password_hash)
		RETURNING id, email, password_hash, paid_at, payment_customer_id, created_at, updated_at
	`, map[string]any{
		":id":            id,
		":email":         email,
		":password_hash": passwordHash,
	}, func(stmt *sqlite.Stmt) (User, error) {
		return scanUser(stmt), nil
	})
	if err != nil {
		if sqlite.ErrCode(err) == sqlite.ResultConstraintUnique && strings.Contains(err.Error(), "users_email_lower_idx") {
			return User{}, ErrDuplicateEmail
		}
		return User{}, fmt.Errorf("insert: %w", err)
	}
	return user, nil
}

func (x *Pool) CreateUsage(ctx context.Context, usage Usage) (Usage, error) {
	id := uuid.New().String()
	messageID := any(nil)
	if usage.MessageID != "" {
		messageID = usage.MessageID
	}

	return execReturning(ctx, x, "insert usage", `
		INSERT INTO usage (
			id, user_id, thread_id, message_id, provider, model,
			input_tokens, cached_input_tokens, output_tokens, reasoning_tokens
		)
		VALUES (
			:id, :user_id, :thread_id, :message_id, :provider, :model,
			:input_tokens, :cached_input_tokens, :output_tokens, :reasoning_tokens
		)
		RETURNING id, user_id, thread_id, message_id, provider, model,
			input_tokens, cached_input_tokens, output_tokens, reasoning_tokens,
			created_at
	`, map[string]any{
		":id":                  id,
		":user_id":             usage.UserID,
		":thread_id":           usage.ThreadID,
		":message_id":          messageID,
		":provider":            usage.Provider,
		":model":               usage.Model,
		":input_tokens":        usage.InputTokens,
		":cached_input_tokens": usage.CachedInputTokens,
		":output_tokens":       usage.OutputTokens,
		":reasoning_tokens":    usage.ReasoningTokens,
	}, func(stmt *sqlite.Stmt) (Usage, error) {
		return scanUsage(stmt), nil
	})
}

func (x *Pool) ListUsageByThread(ctx context.Context, userID, threadID string) ([]Usage, error) {
	return queryAll(ctx, x, "query usage", `
		SELECT usage.id, usage.user_id, usage.thread_id, usage.message_id, usage.provider, usage.model,
		       usage.input_tokens, usage.cached_input_tokens, usage.output_tokens, usage.reasoning_tokens,
		       usage.created_at
		FROM usage
		JOIN threads ON threads.id = usage.thread_id
		WHERE usage.thread_id = :thread_id AND threads.user_id = :user_id
		ORDER BY usage.created_at ASC
	`, map[string]any{":user_id": userID, ":thread_id": threadID}, func(stmt *sqlite.Stmt) (Usage, error) {
		return scanUsage(stmt), nil
	})
}
