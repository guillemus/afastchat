package db

import (
	"context"
	"errors"

	"afastchat.com/polar"
	"github.com/google/uuid"
	"zombiezen.com/go/sqlite"
)

func (x *Pool) GrantLifetimeLicense(ctx context.Context, userID, sourcePaymentID, paymentCustomerID string) error {
	user, err := x.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	if paymentCustomerID == "" {
		paymentCustomerID = user.PaymentCustomerID
	}

	if err := x.exec(ctx, "update payment customer", `
		UPDATE users
		SET payment_customer_id = :payment_customer_id, updated_at = datetime('now')
		WHERE id = :id
		RETURNING id
	`, map[string]any{
		":id":                  userID,
		":payment_customer_id": paymentCustomerID,
	}); err != nil {
		return err
	}

	return x.exec(ctx, "insert lifetime license", `
		INSERT OR IGNORE INTO licenses (id, user_id, source_payment_id, tier)
		VALUES (:id, :user_id, :source_payment_id, :tier)
	`, map[string]any{
		":id":                uuid.New().String(),
		":user_id":           userID,
		":source_payment_id": nullableString(sourcePaymentID),
		":tier":              LicenseTierLifetime,
	})
}

func (x *Pool) UserHasLicense(ctx context.Context, userID string, tier LicenseTier) (bool, error) {
	_, err := queryOne(ctx, x, "query user license", `
		SELECT id, user_id, source_payment_id, tier, granted_at, created_at, updated_at
		FROM licenses
		WHERE user_id = :user_id AND tier = :tier
	`, map[string]any{":user_id": userID, ":tier": tier}, func(stmt *sqlite.Stmt) (License, error) {
		return scanLicense(stmt), nil
	})
	if errors.Is(err, ErrNotFound) {
		return false, nil
	}
	return err == nil, err
}

func (x *Pool) CreatePayment(ctx context.Context, payment Payment) (Payment, error) {
	id := uuid.New().String()
	return execReturning(ctx, x, "insert payment", `
		INSERT INTO payments (
			id, user_id, provider, provider_checkout_id, type, status, amount_cents, credit_cents
		)
		VALUES (
			:id, :user_id, :provider, :provider_checkout_id, :type, :status, :amount_cents, :credit_cents
		)
		RETURNING id, user_id, provider, provider_checkout_id, type, status, amount_cents, credit_cents, created_at, updated_at
	`, map[string]any{
		":id":                   id,
		":user_id":              payment.UserID,
		":provider":             PaymentProviderPolar,
		":provider_checkout_id": payment.ProviderCheckoutID,
		":type":                 payment.Type,
		":status":               payment.Status,
		":amount_cents":         payment.AmountCents,
		":credit_cents":         payment.CreditCents,
	}, func(stmt *sqlite.Stmt) (Payment, error) {
		return scanPayment(stmt), nil
	})
}

func (x *Pool) GetPaymentByCheckoutID(ctx context.Context, providerCheckoutID string) (Payment, error) {
	return queryOne(ctx, x, "query payment", `
		SELECT id, user_id, provider, provider_checkout_id, type, status, amount_cents, credit_cents, created_at, updated_at
		FROM payments
		WHERE provider = :provider AND provider_checkout_id = :provider_checkout_id
	`, map[string]any{":provider": PaymentProviderPolar, ":provider_checkout_id": providerCheckoutID}, func(stmt *sqlite.Stmt) (Payment, error) {
		return scanPayment(stmt), nil
	})
}

func (x *Pool) UpdatePaymentStatus(ctx context.Context, id string, status PaymentStatus) error {
	return x.exec(ctx, "update payment", `
		UPDATE payments
		SET status = :status, updated_at = datetime('now')
		WHERE id = :id
		RETURNING id
	`, map[string]any{":id": id, ":status": status})
}

func (x *Pool) CreatePaymentEvent(ctx context.Context, event PaymentEvent) (PaymentEvent, error) {
	id := uuid.New().String()
	created, err := execReturning(ctx, x, "insert payment event", `
		INSERT OR IGNORE INTO payment_events (id, provider, provider_event_id, type, payload)
		VALUES (:id, :provider, :provider_event_id, :type, :payload)
		RETURNING id, provider, provider_event_id, type, payload, processed_at, created_at
	`, map[string]any{
		":id":                id,
		":provider":          PaymentProviderPolar,
		":provider_event_id": event.ProviderEventID,
		":type":              event.Type,
		":payload":           event.Payload,
	}, func(stmt *sqlite.Stmt) (PaymentEvent, error) {
		return scanPaymentEvent(stmt), nil
	})
	if errors.Is(err, ErrNotFound) {
		return PaymentEvent{}, ErrDuplicatePaymentEvent
	}
	return created, err
}

func (x *Pool) MarkPaymentEventProcessed(ctx context.Context, id string) error {
	return x.exec(ctx, "mark payment event processed", `
		UPDATE payment_events
		SET processed_at = datetime('now')
		WHERE id = :id
	`, map[string]any{":id": id})
}

func (x *Pool) CreateBillingLedgerEntry(ctx context.Context, entry BillingLedgerEntry) (BillingLedgerEntry, error) {
	id := uuid.New().String()
	return execReturning(ctx, x, "insert billing ledger", `
		INSERT INTO billing_ledger (id, user_id, payment_checkout_id, usage_id, type, amount_cents)
		VALUES (:id, :user_id, :payment_checkout_id, :usage_id, :type, :amount_cents)
		RETURNING id, user_id, payment_checkout_id, usage_id, type, amount_cents, created_at
	`, map[string]any{
		":id":                  id,
		":user_id":             entry.UserID,
		":payment_checkout_id": nullableString(entry.PaymentCheckoutID),
		":usage_id":            nullableString(entry.UsageID),
		":type":                entry.Type,
		":amount_cents":        entry.AmountCents,
	}, func(stmt *sqlite.Stmt) (BillingLedgerEntry, error) {
		return scanBillingLedgerEntry(stmt), nil
	})
}

func (x *Pool) CreateInitialCreditIfMissing(ctx context.Context, userID, paymentCheckoutID string, amountCents polar.Cents) error {
	return x.createPaymentCreditIfMissing(ctx, userID, paymentCheckoutID, BillingLedgerTypeInitialCredit, amountCents)
}

func (x *Pool) CreateTopupCreditIfMissing(ctx context.Context, userID, paymentCheckoutID string, amountCents polar.Cents) error {
	return x.createPaymentCreditIfMissing(ctx, userID, paymentCheckoutID, BillingLedgerTypeTopupCredit, amountCents)
}

func (x *Pool) createPaymentCreditIfMissing(ctx context.Context, userID, paymentCheckoutID string, typ BillingLedgerType, amountCents polar.Cents) error {
	return x.exec(ctx, "insert payment credit", `
		INSERT OR IGNORE INTO billing_ledger (id, user_id, payment_checkout_id, type, amount_cents)
		VALUES (:id, :user_id, :payment_checkout_id, :type, :amount_cents)
	`, map[string]any{
		":id":                  uuid.New().String(),
		":user_id":             userID,
		":payment_checkout_id": paymentCheckoutID,
		":type":                typ,
		":amount_cents":        amountCents,
	})
}

func (x *Pool) GetBillingBalanceCents(ctx context.Context, userID string) (polar.Cents, error) {
	return queryOne(ctx, x, "query billing balance", `
		SELECT COALESCE(SUM(amount_cents), 0) AS balance_cents
		FROM billing_ledger
		WHERE user_id = :user_id
	`, map[string]any{":user_id": userID}, func(stmt *sqlite.Stmt) (polar.Cents, error) {
		return polar.Cents(stmt.GetInt64("balance_cents")), nil
	})
}
