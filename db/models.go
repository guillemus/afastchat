package db

import (
	"time"
)

type Folder struct {
	ID        string
	UserID    string
	ParentID  *string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Thread struct {
	ID          string
	UserID      string
	FolderID    *string
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
	PaymentCustomerID string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type LicenseTier string

const (
	LicenseTierLifetime LicenseTier = "lifetime"
)

type License struct {
	ID              string
	UserID          string
	SourcePaymentID string
	Tier            LicenseTier
	GrantedAt       time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
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

const PaymentProviderPolar = "polar"

type PaymentType string

const (
	PaymentTypeLifetime PaymentType = "lifetime"
	PaymentTypeTopup    PaymentType = "topup"
)

type PaymentStatus string

const (
	PaymentStatusCreated   PaymentStatus = "created"
	PaymentStatusSucceeded PaymentStatus = "succeeded"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusExpired   PaymentStatus = "expired"
)

type Payment struct {
	ID                 string
	UserID             string
	Provider           string
	ProviderCheckoutID string
	Type               PaymentType
	Status             PaymentStatus
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

type BillingLedgerType string

const (
	BillingLedgerTypeInitialCredit BillingLedgerType = "initial_credit"
	BillingLedgerTypeTopupCredit   BillingLedgerType = "topup_credit"
	BillingLedgerTypeUsageDebit    BillingLedgerType = "usage_debit"
	BillingLedgerTypeAdjustment    BillingLedgerType = "adjustment"
)

type BillingLedgerEntry struct {
	ID                string
	UserID            string
	PaymentCheckoutID string
	UsageID           string
	Type              BillingLedgerType
	AmountCents       int64
	CreatedAt         time.Time
}
