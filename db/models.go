package db

import (
	"time"
)

// Thread is the public domain model for a conversation thread
type Thread struct {
	ID          string
	Title       string
	ArchivedAt  *time.Time // nil if not archived
	IsStreaming bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Message is the public domain model for a chat message
type Message struct {
	ID        string
	ThreadID  string
	Role      string // 'user', 'assistant', 'system'
	Content   string
	CreatedAt time.Time
}
