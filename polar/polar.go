package polar

import polargo "github.com/polarsource/polar-go"

type Polar struct {
	Client *polargo.Polar
}

func New(apiKey string) *Polar {
	return &Polar{
		Client: polargo.New(
			polargo.WithSecurity(apiKey),
		),
	}
}
