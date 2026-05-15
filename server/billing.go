package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"afastchat.com/db"
	"afastchat.com/env"
	"afastchat.com/polar"
	"afastchat.com/views"

	"github.com/polarsource/polar-go/models/components"
	"github.com/starfederation/datastar-go/datastar"
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

type BalanceCheckoutSignals struct {
	AmountDollars string `json:"amountDollars"`
}

func (x *Server) handleBalanceCheckout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := mustGetUser(r)
	var sig BalanceCheckoutSignals
	if err := datastar.ReadSignals(r, &sig); err != nil {
		toastShowError(w, r, "Unable to read amount", "Refresh the page and enter a whole dollar amount.")
		return
	}
	dollars, err := strconv.ParseInt(sig.AmountDollars, 10, 64)
	if err != nil {
		toastShowError(w, r, "Enter a whole dollar amount", "Use dollars only, with no cents or decimals.")
		return
	}
	amount := dollars * 100
	amountCents := polar.Cents(amount)
	if amountCents < polar.ProductBalanceMinimumCents || amountCents > polar.ProductBalanceMaximumCents {
		msg := fmt.Sprintf(
			"Enter an amount between %s and %s.",
			polar.ProductBalanceMinimumCents.Format(), polar.ProductBalanceMaximumCents.Format(),
		)
		toastShowError(w, r, "Amount out of range", msg)
		return
	}

	baseURL := env.AppURL(x.env)
	successURL := baseURL + "/settings/thank-you?checkout_id={CHECKOUT_ID}"
	returnURL := baseURL + "/settings"

	res, err := x.polar.Client.Checkouts.Create(ctx, components.CheckoutCreate{
		Products:           []string{polar.ProductIDBalance},
		ExternalCustomerID: &user.ID,
		CustomerEmail:      &user.Email,
		SuccessURL:         &successURL,
		ReturnURL:          &returnURL,
		Amount:             &amount,
		Metadata: map[string]components.CheckoutCreateMetadata{
			"type":         components.CreateCheckoutCreateMetadataStr(string(db.PaymentTypeTopup)),
			"user_id":      components.CreateCheckoutCreateMetadataStr(user.ID),
			"credit_cents": components.CreateCheckoutCreateMetadataInteger(amount),
		},
	})
	if err != nil {
		slog.Error("create balance checkout", "err", err)
		toastShowError(w, r,
			"Unable to add usage credit",
			"We couldn't start checkout. No payment was created. Try again.")
		return
	}

	_, err = x.db.CreatePayment(ctx, db.Payment{
		UserID:             user.ID,
		ProviderCheckoutID: res.Checkout.ID,
		Type:               db.PaymentTypeTopup,
		Status:             db.PaymentStatusCreated,
		AmountCents:        amount,
		CreditCents:        amount,
	})
	if err != nil {
		slog.Error("record balance checkout", "err", err)
		toastShowError(w, r,
			"Unable to add usage credit",
			"We couldn't save the checkout on our side. Do not pay from this checkout link; try again.")
		return
	}

	sse := datastar.NewSSE(w, r)
	sse.Redirect(res.Checkout.URL)
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
	creditCents := polar.Cents(0)
	paymentType := db.PaymentType("")
	res, err := x.polar.Client.Checkouts.Get(ctx, checkoutID)
	if err != nil {
		slog.Error("get polar checkout", "err", err, "checkout_id", checkoutID)
	} else if res.Checkout.Status == components.CheckoutStatusSucceeded {
		checkout, err := x.db.GetPaymentByCheckoutID(ctx, checkoutID)
		if err != nil {
			handleBadReq(w, "get payment checkout", err)
			return
		}
		creditCents = polar.Cents(checkout.CreditCents)
		paymentType = checkout.Type
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
		if checkout.Type == db.PaymentTypeLifetime {
			if err := x.db.MarkUserPaid(ctx, user.ID, customerID); err != nil {
				handleBadReq(w, "mark user paid", err)
				return
			}
			if err := x.db.CreateInitialCreditIfMissing(ctx, user.ID, checkout.ID, polar.ProductProIncludedCreditCents); err != nil {
				handleBadReq(w, "create initial credit", err)
				return
			}
		}
		if checkout.Type == db.PaymentTypeTopup {
			if err := x.db.CreateTopupCreditIfMissing(ctx, user.ID, checkout.ID, polar.Cents(checkout.CreditCents)); err != nil {
				handleBadReq(w, "create topup credit", err)
				return
			}
		}
		confirmed = true
	}

	x.renderer.RenderCheckoutThankYouPage(w, views.CheckoutThankYouData{
		Confirmed:   confirmed,
		PaymentType: string(paymentType),
		CreditCents: creditCents,
	})
}
