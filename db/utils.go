package db

import (
	"fmt"
	"log/slog"
	"time"

	"zombiezen.com/go/sqlite"
)

func parseTime(v string) time.Time {
	if v == "" {
		return time.Time{}
	}
	// SQLite DATETIME stores as TEXT in ISO8601 format: "YYYY-MM-DD HH:MM:SS"
	parsed, err := time.Parse("2006-01-02 15:04:05", v)
	if err != nil {
		slog.Error("datetime parse error", "err", err, "value", v)
		return time.Time{}
	}
	return parsed
}

func scanThread(stmt *sqlite.Stmt) Thread {
	return Thread{
		ID:          stmt.GetText("id"),
		Title:       stmt.GetText("title"),
		ArchivedAt:  nullableTime(stmt.IsNull("archived_at"), stmt.GetText("archived_at")),
		IsStreaming: stmt.GetInt64("is_streaming") != 0,
		CreatedAt:   parseTime(stmt.GetText("created_at")),
		UpdatedAt:   parseTime(stmt.GetText("updated_at")),
	}
}

func scanMessage(stmt *sqlite.Stmt) (Message, error) {
	content, err := scanMessageContent(stmt)
	if err != nil {
		return Message{}, fmt.Errorf("scan content: %w", err)
	}

	return Message{
		ID:                 stmt.GetText("id"),
		ThreadID:           stmt.GetText("thread_id"),
		Type:               stmt.GetText("type"),
		Seq:                int(stmt.GetInt64("seq")),
		Content:            content,
		ToolCallID:         stmt.GetText("tool_call_id"),
		ToolName:           stmt.GetText("tool_name"),
		ProviderResponseID: stmt.GetText("provider_response_id"),
		CreatedAt:          parseTime(stmt.GetText("created_at")),
		UpdatedAt:          parseTime(stmt.GetText("updated_at")),
	}, nil
}

func nullableTime(isNull bool, val string) *time.Time {
	if isNull || val == "" {
		return nil
	}
	t := parseTime(val)
	return &t
}
