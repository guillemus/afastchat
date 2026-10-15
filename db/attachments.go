package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type Attachment struct {
	ID        string
	MessageID string
	ObjectKey string
	MimeType  string
	Filename  string
	URL       string
}

type AttachmentInput struct {
	ObjectKey string
	MimeType  string
	Filename  string
}

func (x *Pool) CreateMessageAttachments(ctx context.Context, userID, messageID string, inputs []AttachmentInput) error {
	if len(inputs) == 0 {
		return nil
	}

	return x.WithConn(ctx, func(conn *sqlite.Conn) (err error) {
		defer sqlitex.Save(conn)(&err)

		for position, input := range inputs {
			err := sqlitex.Execute(conn, `
				INSERT INTO message_attachments (id, message_id, user_id, object_key, mime_type, filename, position)
				VALUES (:id, :message_id, :user_id, :object_key, :mime_type, :filename, :position)
			`, &sqlitex.ExecOptions{Named: map[string]any{
				":id":         uuid.New().String(),
				":message_id": messageID,
				":user_id":    userID,
				":object_key": input.ObjectKey,
				":mime_type":  input.MimeType,
				":filename":   input.Filename,
				":position":   position,
			}})
			if err != nil {
				return fmt.Errorf("create message attachment: %w", err)
			}
		}
		return nil
	})
}

func (x *Pool) ListThreadAttachments(ctx context.Context, userID, threadID string) (map[string][]Attachment, error) {
	attachments, err := queryAll(ctx, x, "query thread attachments", `
		SELECT message_attachments.id, message_attachments.message_id, message_attachments.object_key,
		       message_attachments.mime_type, message_attachments.filename
		FROM message_attachments
		JOIN messages ON messages.id = message_attachments.message_id
		WHERE messages.thread_id = :thread_id AND message_attachments.user_id = :user_id
		ORDER BY message_attachments.message_id, message_attachments.position
	`, map[string]any{":thread_id": threadID, ":user_id": userID}, func(stmt *sqlite.Stmt) (Attachment, error) {
		return Attachment{
			ID:        stmt.GetText("id"),
			MessageID: stmt.GetText("message_id"),
			ObjectKey: stmt.GetText("object_key"),
			MimeType:  stmt.GetText("mime_type"),
			Filename:  stmt.GetText("filename"),
		}, nil
	})
	if err != nil {
		return nil, err
	}

	byMessage := make(map[string][]Attachment)
	for _, attachment := range attachments {
		byMessage[attachment.MessageID] = append(byMessage[attachment.MessageID], attachment)
	}
	return byMessage, nil
}

func selectAttachmentKeys(conn *sqlite.Conn, query string, named map[string]any) ([]string, error) {
	var keys []string
	err := sqlitex.Execute(conn, query, &sqlitex.ExecOptions{
		Named: named,
		ResultFunc: func(stmt *sqlite.Stmt) error {
			keys = append(keys, stmt.GetText("object_key"))
			return nil
		},
	})
	if err != nil {
		return nil, fmt.Errorf("select attachment keys: %w", err)
	}
	return keys, nil
}
