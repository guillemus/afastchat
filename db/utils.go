package db

import (
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

func scanMessage(stmt *sqlite.Stmt) Message {
	return Message{
		ID:        stmt.GetText("id"),
		ThreadID:  stmt.GetText("thread_id"),
		Role:      stmt.GetText("role"),
		Content:   stmt.GetText("content"),
		CreatedAt: parseTime(stmt.GetText("created_at")),
	}
}

func nullableTime(isNull bool, val string) *time.Time {
	if isNull || val == "" {
		return nil
	}
	t := parseTime(val)
	return &t
}
