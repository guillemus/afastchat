package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"afastchat.com/db"
	"afastchat.com/env"
	"afastchat.com/polar"

	"github.com/polarsource/polar-go/models/components"
)

func (x *Server) handleProCheckout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := mustGetUser(r)
	baseURL := env.AppURL(x.env)
	successURL := baseURL + "/settings/thank-you?checkout_id={CHECKOUT_ID}"
	returnURL := baseURL + "/settings"

	res, err := x.polar.Client.Checkouts.Create(ctx, components.CheckoutCreate{
		Products:           []string{polar.ProductIDPro},
		ExternalCustomerID: &user.ID,
		CustomerEmail:      &user.Email,
		SuccessURL:         &successURL,
		ReturnURL:          &returnURL,
		Metadata: map[string]components.CheckoutCreateMetadata{
			"type":    components.CreateCheckoutCreateMetadataStr(string(db.PaymentTypeLifetime)),
			"user_id": components.CreateCheckoutCreateMetadataStr(user.ID),
		},
	})
	if err != nil {
		handleBadReq(w, "create polar checkout", err)
		return
	}

	_, err = x.db.CreatePayment(ctx, db.Payment{
		UserID:             user.ID,
		ProviderCheckoutID: res.Checkout.ID,
		Type:               db.PaymentTypeLifetime,
		Status:             db.PaymentStatusCreated,
		AmountCents:        int64(polar.ProductProPriceCents),
		CreditCents:        int64(polar.ProductProIncludedCreditCents),
	})
	if err != nil {
		handleBadReq(w, "record payment checkout", err)
		return
	}

	http.Redirect(w, r, res.Checkout.URL, http.StatusSeeOther)
}

func (x *Server) handleCheckoutThankYou(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := mustGetUser(r)
	checkoutID := r.URL.Query().Get("checkout_id")
	if checkoutID == "" {
		handleBadReq(w, "missing checkout_id", fmt.Errorf("missing checkout_id"))
		return
	}

	confirmed := false
	res, err := x.polar.Client.Checkouts.Get(ctx, checkoutID)
	if err != nil {
		slog.Error("get polar checkout", "err", err, "checkout_id", checkoutID)
	} else if res.Checkout.Status == components.CheckoutStatusSucceeded {
		checkout, err := x.db.GetPaymentByCheckoutID(ctx, checkoutID)
		if err != nil {
			handleBadReq(w, "get payment checkout", err)
			return
		}
		if checkout.UserID != user.ID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		if err := x.db.UpdatePaymentStatus(ctx, checkout.ID, db.PaymentStatusSucceeded); err != nil {
			handleBadReq(w, "update payment checkout", err)
			return
		}

		customerID := ""
		if res.Checkout.CustomerID != nil {
			customerID = *res.Checkout.CustomerID
		}
		if err := x.db.MarkUserPaid(ctx, user.ID, customerID); err != nil {
			handleBadReq(w, "mark user paid", err)
			return
		}

		if err := x.db.CreateInitialCreditIfMissing(ctx, user.ID, checkout.ID, polar.ProductProIncludedCreditCents); err != nil {
			handleBadReq(w, "create initial credit", err)
			return
		}
		confirmed = true
	}

	x.renderer.RenderCheckoutThankYouPage(w, confirmed)
}
