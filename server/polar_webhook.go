package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"afastchat.com/db"
	"afastchat.com/polar"

	standardwebhooks "github.com/standard-webhooks/standard-webhooks/libraries/go"
)

type polarWebhookEvent struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Metadata struct {
		OrderID string `json:"order_id"`
	} `json:"metadata"`
}

func (x *Server) handlePolarWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	secret := base64.StdEncoding.EncodeToString([]byte(x.cfg.PolarWebhookSecret))
	wh, err := standardwebhooks.NewWebhook(secret)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	if err := wh.Verify(body, r.Header); err != nil {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	var event polarWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	logger := slog.With("event_id", event.ID, "event", event.Name)

	saved, err := x.db.CreatePaymentEvent(ctx, db.PaymentEvent{
		ProviderEventID: event.ID,
		Type:            event.Name,
		Payload:         string(body),
	})
	if errors.Is(err, db.ErrDuplicatePaymentEvent) {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	if err != nil {
		logger.Error("save polar webhook", "err", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := x.processPolarWebhook(ctx, event); err != nil {
			logger.Error("process polar webhook", "err", err)
			return
		}

		if err := x.db.MarkPaymentEventProcessed(ctx, saved.ID); err != nil {
			logger.Error("mark polar webhook processed", "err", err)
			return
		}
	}()
}

func (x *Server) processPolarWebhook(ctx context.Context, event polarWebhookEvent) error {
	switch event.Name {
	case "order.paid":
		return x.processPolarOrderPaid(ctx, event)
	default:
		return nil
	}
}

func (x *Server) processPolarOrderPaid(ctx context.Context, event polarWebhookEvent) error {
	res, err := x.polar.Client.Orders.Get(ctx, event.Metadata.OrderID)
	if err != nil {
		return fmt.Errorf("get polar order: %w", err)
	}
	order := res.Order
	if order.CheckoutID == nil {
		return nil
	}

	payment, err := x.db.GetPaymentByCheckoutID(ctx, *order.CheckoutID)
	if err != nil {
		return fmt.Errorf("get payment: %w", err)
	}
	if err := x.db.UpdatePaymentStatus(ctx, payment.ID, db.PaymentStatusSucceeded); err != nil {
		return fmt.Errorf("update payment: %w", err)
	}
	if payment.Type == db.PaymentTypeLifetime {
		if err := x.db.GrantLifetimeLicense(ctx, payment.UserID, payment.ID, order.CustomerID); err != nil {
			return fmt.Errorf("grant lifetime license: %w", err)
		}
		if err := x.db.CreateInitialCreditIfMissing(ctx, payment.UserID, payment.ID, polar.ProductProIncludedCreditCents); err != nil {
			return fmt.Errorf("create initial credit: %w", err)
		}
	}
	if payment.Type == db.PaymentTypeTopup {
		if err := x.db.CreateTopupCreditIfMissing(ctx, payment.UserID, payment.ID, polar.Cents(payment.CreditCents)); err != nil {
			return fmt.Errorf("create topup credit: %w", err)
		}
	}
	return nil
}
