package db

import (
	"time"
)

type Thread struct {
	ID          string
	UserID      string
	Title       string
	ArchivedAt  *time.Time // nil if not archived
	IsStreaming bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type MessageType string

const (
	UserMessage        MessageType = "user_message"
	AssistantMessage   MessageType = "assistant_message"
	SystemMessage      MessageType = "system_message"
	AssistantReasoning MessageType = "assistant_reasoning"
	AssistantToolCall  MessageType = "assistant_tool_call"
	ToolOutput         MessageType = "tool_output"
)

type Message struct {
	ID                 string
	ThreadID           string
	Type               MessageType
	Seq                int
	Content            string
	ToolCallID         string
	ToolName           string
	ProviderResponseID string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type User struct {
	ID                string
	Email             string
	PasswordHash      string
	PaidAt            *time.Time
	PaymentCustomerID string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Usage struct {
	ID                string
	UserID            string
	ThreadID          string
	MessageID         string
	Provider          string
	Model             string
	InputTokens       int64
	CachedInputTokens int64
	OutputTokens      int64
	ReasoningTokens   int64
	CreatedAt         time.Time
}

type Payment struct {
	ID                 string
	UserID             string
	Provider           string
	ProviderCheckoutID string
	Type               string
	Status             string
	AmountCents        int64
	CreditCents        int64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type PaymentEvent struct {
	ID              string
	Provider        string
	ProviderEventID string
	Type            string
	Payload         string
	ProcessedAt     *time.Time
	CreatedAt       time.Time
}

type BillingLedgerEntry struct {
	ID                string
	UserID            string
	PaymentCheckoutID string
	UsageID           string
	Type              string
	AmountCents       int64
	CreatedAt         time.Time
}
