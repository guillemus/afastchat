package polar

import (
	"fmt"

	polargo "github.com/polarsource/polar-go"
)

type Polar struct {
	Client *polargo.Polar
}

const (
	// ProductIDBalance is the usage credit top-up product.
	ProductIDBalance = "3868e797-ffcb-4da9-8651-5b780b64975e"

	ProductBalanceMinimumCents Cents = 500
	ProductBalanceMaximumCents Cents = 10000

	// ProductIDPro is the lifetime access product.
	ProductIDPro = "79bec933-c575-4b3b-9963-d9cfafa9e194"

	ProductProPriceCents          Cents = 1700
	ProductProIncludedCreditCents Cents = 500
)

type Cents int64

func (x Cents) Format() string {
	return fmt.Sprintf("$%d.%02d", x/100, x%100)
}

func New(apiKey string) *Polar {
	return &Polar{
		Client: polargo.New(
			polargo.WithSecurity(apiKey),
		),
	}
}
