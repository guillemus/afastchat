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

func nullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func scanThread(stmt *sqlite.Stmt) Thread {
	return Thread{
		ID:          stmt.GetText("id"),
		UserID:      stmt.GetText("user_id"),
		Title:       stmt.GetText("title"),
		ArchivedAt:  nullableTime(stmt.IsNull("archived_at"), stmt.GetText("archived_at")),
		IsStreaming: stmt.GetInt64("is_streaming") != 0,
		CreatedAt:   parseTime(stmt.GetText("created_at")),
		UpdatedAt:   parseTime(stmt.GetText("updated_at")),
	}
}

func scanMessage(stmt *sqlite.Stmt) (Message, error) {
	return Message{
		ID:                 stmt.GetText("id"),
		ThreadID:           stmt.GetText("thread_id"),
		Type:               MessageType(stmt.GetText("type")),
		Seq:                int(stmt.GetInt64("seq")),
		Content:            stmt.GetText("content"),
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

func scanUser(stmt *sqlite.Stmt) User {
	return User{
		ID:                stmt.GetText("id"),
		Email:             stmt.GetText("email"),
		PasswordHash:      stmt.GetText("password_hash"),
		PaidAt:            nullableTime(stmt.IsNull("paid_at"), stmt.GetText("paid_at")),
		PaymentCustomerID: stmt.GetText("payment_customer_id"),
		CreatedAt:         parseTime(stmt.GetText("created_at")),
		UpdatedAt:         parseTime(stmt.GetText("updated_at")),
	}
}

func scanUsage(stmt *sqlite.Stmt) Usage {
	return Usage{
		ID:                stmt.GetText("id"),
		UserID:            stmt.GetText("user_id"),
		ThreadID:          stmt.GetText("thread_id"),
		MessageID:         stmt.GetText("message_id"),
		Provider:          stmt.GetText("provider"),
		Model:             stmt.GetText("model"),
		InputTokens:       stmt.GetInt64("input_tokens"),
		CachedInputTokens: stmt.GetInt64("cached_input_tokens"),
		OutputTokens:      stmt.GetInt64("output_tokens"),
		ReasoningTokens:   stmt.GetInt64("reasoning_tokens"),
		CreatedAt:         parseTime(stmt.GetText("created_at")),
	}
}

func scanPayment(stmt *sqlite.Stmt) Payment {
	return Payment{
		ID:                 stmt.GetText("id"),
		UserID:             stmt.GetText("user_id"),
		Provider:           stmt.GetText("provider"),
		ProviderCheckoutID: stmt.GetText("provider_checkout_id"),
		Type:               stmt.GetText("type"),
		Status:             stmt.GetText("status"),
		AmountCents:        stmt.GetInt64("amount_cents"),
		CreditCents:        stmt.GetInt64("credit_cents"),
		CreatedAt:          parseTime(stmt.GetText("created_at")),
		UpdatedAt:          parseTime(stmt.GetText("updated_at")),
	}
}

func scanPaymentEvent(stmt *sqlite.Stmt) PaymentEvent {
	return PaymentEvent{
		ID:              stmt.GetText("id"),
		Provider:        stmt.GetText("provider"),
		ProviderEventID: stmt.GetText("provider_event_id"),
		Type:            stmt.GetText("type"),
		Payload:         stmt.GetText("payload"),
		ProcessedAt:     nullableTime(stmt.IsNull("processed_at"), stmt.GetText("processed_at")),
		CreatedAt:       parseTime(stmt.GetText("created_at")),
	}
}

func scanBillingLedgerEntry(stmt *sqlite.Stmt) BillingLedgerEntry {
	return BillingLedgerEntry{
		ID:                stmt.GetText("id"),
		UserID:            stmt.GetText("user_id"),
		PaymentCheckoutID: stmt.GetText("payment_checkout_id"),
		UsageID:           stmt.GetText("usage_id"),
		Type:              stmt.GetText("type"),
		AmountCents:       stmt.GetInt64("amount_cents"),
		CreatedAt:         parseTime(stmt.GetText("created_at")),
	}
}
