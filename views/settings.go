package views

import (
	"fmt"

	"afastchat.com/polar"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type SettingsData struct {
	Threads            []Thread
	UserEmail          string
	HasLifetimeLicense bool
	BalanceCents       polar.Cents
}

func SettingsPage(data SettingsData) Node {
	return Main(Class("flex-1 overflow-y-auto bg-neutral-950 p-8"),
		Div(Class("mx-auto max-w-[760px]"),
			H1(Class("mb-5 text-3xl font-medium tracking-[-0.03em]"), Text("Settings")),
			Section(Class("mb-3 rounded-xl border border-neutral-800 bg-[color-mix(in_oklch,#171717_55%,transparent)] p-5"),
				H2(Class("mb-3 text-base font-medium"), Text("Account")),
				Div(Class("flex items-center justify-between gap-4 py-2 text-neutral-400"),
					Span(Text("Email")),
					Span(Class("font-medium text-neutral-100"), Text(data.UserEmail)),
				),
			),
			If(data.HasLifetimeLicense, BillingCard(data)),
			If(!data.HasLifetimeLicense, PaywallCard()),
		),
	)
}

func BillingCard(data SettingsData) Node {
	return Section(Class("mb-3 rounded-xl border border-neutral-800 bg-[color-mix(in_oklch,#171717_55%,transparent)] p-5"),
		H2(Class("mb-3 text-base font-medium"), Text("Billing")),
		Div(Class("flex items-center justify-between gap-4 py-2 text-neutral-400"),
			Span(Text("Plan")),
			Span(Class("font-medium text-neutral-100"), Text(planLabel(data.HasLifetimeLicense))),
		),
		Div(Class("flex items-center justify-between gap-4 py-2 text-neutral-400"),
			Span(Text("Usage balance")),
			Span(Class("font-medium text-neutral-100"), Text(data.BalanceCents.Format())),
		),
		Div(Class("mt-3 flex items-center gap-2"),
			Data("signals:amount-dollars", `"5"`),
			Label(Class("relative"),
				Span(Class("pointer-events-none absolute left-2 top-1/2 -translate-y-1/2 text-neutral-400"), Text("$")),
				Input(Class("min-h-9 w-[140px] rounded-lg border border-neutral-800 bg-neutral-900 px-2 pl-5 font-[inherit] text-neutral-100"), Type("text"), Value("5"), Attr("inputmode", "numeric"), Pattern("[0-9]*"), Data("bind:amount-dollars", "")),
			),
			Button(Type("button"), Class("inline-flex min-h-9 cursor-pointer items-center justify-center rounded-lg border-0 bg-orange-500 px-3 font-[inherit] text-white no-underline"), Data("on:click", `@post('/settings/checkout/balance')`), Text("Add usage credit")),
		),
	)
}

func PaywallCard() Node {
	return Section(Class("mb-3 rounded-xl border border-neutral-800 bg-[color-mix(in_oklch,#171717_55%,transparent)] p-5"),
		H2(Class("mb-3 text-base font-medium"), Text("Unlock A Fast Chat")),
		P(Class("mb-4 leading-7 text-neutral-400"),
			Text(fmt.Sprintf(
				"Lifetime access is %s and includes %s of usage credit.",
				polar.ProductProPriceCents.Format(), polar.ProductProIncludedCreditCents.Format(),
			)),
		),
		Form(Method("post"), Action("/settings/checkout/pro"),
			Button(Type("submit"), Class("inline-flex min-h-9 cursor-pointer items-center justify-center rounded-lg border-0 bg-orange-500 px-3 font-[inherit] text-white no-underline"), Text("Buy lifetime access")),
		),
	)
}

type CheckoutThankYouData struct {
	Confirmed   bool
	PaymentType string
	CreditCents polar.Cents
}

func CheckoutThankYouPage(data CheckoutThankYouData) Node {
	return Main(Class("flex-1 overflow-y-auto bg-neutral-950 p-8"),
		Div(Class("mx-auto max-w-[760px]"),
			Section(Class("mb-3 rounded-xl border border-neutral-800 bg-[color-mix(in_oklch,#171717_55%,transparent)] p-5"),
				H1(Class("mb-5 text-3xl font-medium tracking-[-0.03em]"), Text(thankYouTitle(data.Confirmed))),
				P(Class("mb-4 leading-7 text-neutral-400"), Text(thankYouCopy(data))),
				A(Class("inline-flex min-h-9 cursor-pointer items-center justify-center rounded-lg border-0 bg-orange-500 px-3 font-[inherit] text-white no-underline"), Href("/"), Text("Start chatting")),
			),
		),
	)
}

func planLabel(hasLifetimeLicense bool) string {
	if hasLifetimeLicense {
		return "Lifetime access"
	}
	return "Not active"
}

func thankYouTitle(confirmed bool) string {
	if confirmed {
		return "Payment confirmed"
	}
	return "Payment pending"
}

func thankYouCopy(data CheckoutThankYouData) string {
	if !data.Confirmed {
		return "We could not confirm the payment yet. Refresh this page in a moment."
	}
	if data.PaymentType == "lifetime" {
		return fmt.Sprintf("Your lifetime access is active and %s of usage credit was added to your account.", data.CreditCents.Format())
	}
	return fmt.Sprintf("%s of usage credit was added to your account.", data.CreditCents.Format())
}
