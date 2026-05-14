package db

import (
	"context"
	"fmt"
	"time"

	"afastchat.com/polar"
	"github.com/google/uuid"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func (x *Pool) MarkUserPaid(ctx context.Context, userID, paymentCustomerID string) error {
	user, err := x.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	paidAt := time.Now().UTC().Format("2006-01-02 15:04:05")
	if user.PaidAt != nil {
		paidAt = user.PaidAt.Format("2006-01-02 15:04:05")
	}

	if paymentCustomerID == "" {
		paymentCustomerID = user.PaymentCustomerID
	}

	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		var updatedID string
		err := sqlitex.Execute(conn, `
			UPDATE users
			SET paid_at = :paid_at, payment_customer_id = :payment_customer_id, updated_at = datetime('now')
			WHERE id = :id
			RETURNING id
		`, &sqlitex.ExecOptions{
			Named: map[string]any{
				":id":                  userID,
				":paid_at":             paidAt,
				":payment_customer_id": paymentCustomerID,
			},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				updatedID = stmt.GetText("id")
				return nil
			},
		})
		if err != nil {
			return fmt.Errorf("update user paid: %w", err)
		}
		if updatedID == "" {
			return ErrNotFound
		}
		return nil
	})
}

func (x *Pool) CreatePayment(
	ctx context.Context, payment Payment,
) (Payment, error) {
	id := uuid.New().String()
	var created Payment
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		err := sqlitex.Execute(conn, `
			INSERT INTO payments (
				id, user_id, provider, provider_checkout_id, type, status, amount_cents, credit_cents
			)
			VALUES (
				:id, :user_id, :provider, :provider_checkout_id, :type, :status, :amount_cents, :credit_cents
			)
			RETURNING id, user_id, provider, provider_checkout_id, type, status, amount_cents, credit_cents, created_at, updated_at
		`, &sqlitex.ExecOptions{
			Named: map[string]any{
				":id":                   id,
				":user_id":              payment.UserID,
				":provider":             payment.Provider,
				":provider_checkout_id": payment.ProviderCheckoutID,
				":type":                 payment.Type,
				":status":               payment.Status,
				":amount_cents":         payment.AmountCents,
				":credit_cents":         payment.CreditCents,
			},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				created = scanPayment(stmt)
				return nil
			},
		})
		if err != nil {
			return fmt.Errorf("insert payment: %w", err)
		}
		return nil
	})
	return created, err
}

func (x *Pool) GetPaymentByProviderCheckoutID(ctx context.Context, provider, providerCheckoutID string) (Payment, error) {
	var payment Payment
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		err := sqlitex.Execute(conn, `
			SELECT id, user_id, provider, provider_checkout_id, type, status, amount_cents, credit_cents, created_at, updated_at
			FROM payments
			WHERE provider = :provider AND provider_checkout_id = :provider_checkout_id
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":provider": provider, ":provider_checkout_id": providerCheckoutID},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				payment = scanPayment(stmt)
				return nil
			},
		})
		if err != nil {
			return fmt.Errorf("query payment: %w", err)
		}
		if payment.ID == "" {
			return ErrNotFound
		}
		return nil
	})
	return payment, err
}

func (x *Pool) UpdatePaymentStatus(ctx context.Context, id, status string) error {
	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		var updatedID string
		err := sqlitex.Execute(conn, `
			UPDATE payments
			SET status = :status, updated_at = datetime('now')
			WHERE id = :id
			RETURNING id
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":id": id, ":status": status},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				updatedID = stmt.GetText("id")
				return nil
			},
		})
		if err != nil {
			return fmt.Errorf("update payment: %w", err)
		}
		if updatedID == "" {
			return ErrNotFound
		}
		return nil
	})
}

func (x *Pool) CreatePaymentEvent(ctx context.Context, event PaymentEvent) (PaymentEvent, error) {
	id := uuid.New().String()
	var created PaymentEvent
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		err := sqlitex.Execute(conn, `
			INSERT OR IGNORE INTO payment_events (id, provider, provider_event_id, type, payload)
			VALUES (:id, :provider, :provider_event_id, :type, :payload)
			RETURNING id, provider, provider_event_id, type, payload, processed_at, created_at
		`, &sqlitex.ExecOptions{
			Named: map[string]any{
				":id":                id,
				":provider":          event.Provider,
				":provider_event_id": event.ProviderEventID,
				":type":              event.Type,
				":payload":           event.Payload,
			},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				created = scanPaymentEvent(stmt)
				return nil
			},
		})
		if err != nil {
			return fmt.Errorf("insert payment event: %w", err)
		}
		if created.ID == "" {
			return ErrDuplicatePaymentEvent
		}
		return nil
	})
	return created, err
}

func (x *Pool) MarkPaymentEventProcessed(ctx context.Context, id string) error {
	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.Execute(conn, `
			UPDATE payment_events
			SET processed_at = datetime('now')
			WHERE id = :id
		`, &sqlitex.ExecOptions{Named: map[string]any{":id": id}})
	})
}

func (x *Pool) CreateBillingLedgerEntry(ctx context.Context, entry BillingLedgerEntry) (BillingLedgerEntry, error) {
	id := uuid.New().String()
	var created BillingLedgerEntry
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		err := sqlitex.Execute(conn, `
			INSERT INTO billing_ledger (id, user_id, payment_checkout_id, usage_id, type, amount_cents)
			VALUES (:id, :user_id, :payment_checkout_id, :usage_id, :type, :amount_cents)
			RETURNING id, user_id, payment_checkout_id, usage_id, type, amount_cents, created_at
		`, &sqlitex.ExecOptions{
			Named: map[string]any{
				":id":                  id,
				":user_id":             entry.UserID,
				":payment_checkout_id": nullableString(entry.PaymentCheckoutID),
				":usage_id":            nullableString(entry.UsageID),
				":type":                entry.Type,
				":amount_cents":        entry.AmountCents,
			},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				created = scanBillingLedgerEntry(stmt)
				return nil
			},
		})
		if err != nil {
			return fmt.Errorf("insert billing ledger: %w", err)
		}
		return nil
	})
	return created, err
}

func (x *Pool) CreateInitialCreditIfMissing(ctx context.Context, userID, paymentCheckoutID string, amountCents polar.Cents) error {
	return x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.Execute(conn, `
			INSERT OR IGNORE INTO billing_ledger (id, user_id, payment_checkout_id, type, amount_cents)
			VALUES (:id, :user_id, :payment_checkout_id, 'initial_credit', :amount_cents)
		`, &sqlitex.ExecOptions{
			Named: map[string]any{
				":id":                  uuid.New().String(),
				":user_id":             userID,
				":payment_checkout_id": paymentCheckoutID,
				":amount_cents":        amountCents,
			},
		})
	})
}

func (x *Pool) GetBillingBalanceCents(ctx context.Context, userID string) (polar.Cents, error) {
	var balance polar.Cents
	err := x.WithConn(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.Execute(conn, `
			SELECT COALESCE(SUM(amount_cents), 0) AS balance_cents
			FROM billing_ledger
			WHERE user_id = :user_id
		`, &sqlitex.ExecOptions{
			Named: map[string]any{":user_id": userID},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				balance = polar.Cents(stmt.GetInt64("balance_cents"))
				return nil
			},
		})
	})
	return balance, err
}
