package views

import (
	"fmt"

	"afastchat.com/polar"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type SettingsData struct {
	Threads      []Thread
	UserEmail    string
	Paid         bool
	BalanceCents polar.Cents
}

var _ = styles.Style(`
	.settings-page {
		flex: 1;
		overflow-y: auto;
		padding: var(--size-8);
		background: var(--bg-primary);
	}
	.settings-shell {
		max-width: 760px;
		margin: 0 auto;
	}
	.settings-title {
		font-size: var(--font-size-5);
		font-weight: var(--font-weight-5);
		letter-spacing: -0.03em;
		margin-bottom: var(--size-5);
	}
	.settings-card {
		border: 1px solid var(--border-color);
		border-radius: var(--radius-3);
		background: color-mix(in oklch, var(--gray-11) 55%, transparent);
		padding: var(--size-5);
		margin-bottom: var(--size-3);
	}
	.settings-card-title {
		font-size: var(--font-size-2);
		font-weight: var(--font-weight-5);
		margin-bottom: var(--size-3);
	}
	.settings-copy {
		color: var(--text-secondary);
		line-height: var(--font-lineheight-3);
		margin-bottom: var(--size-4);
	}
	.settings-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--size-4);
		padding: var(--size-2) 0;
		color: var(--text-secondary);
	}
	.settings-value {
		color: var(--text-primary);
		font-weight: var(--font-weight-5);
	}
	.settings-action {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-height: 36px;
		padding: 0 var(--size-3);
		border-radius: var(--radius-2);
		background: var(--bg-button);
		color: white;
		border: 0;
		font: inherit;
		text-decoration: none;
		cursor: pointer;
	}
	.sidebar-footer-link {
		display: flex;
		align-items: center;
		min-width: 0;
		text-decoration: none;
	}
`)

func SettingsPage(data SettingsData) Node {
	return Main(Class("settings-page"),
		Div(Class("settings-shell"),
			H1(Class("settings-title"), Text("Settings")),
			Section(Class("settings-card"),
				H2(Class("settings-card-title"), Text("Account")),
				Div(Class("settings-row"),
					Span(Text("Email")),
					Span(Class("settings-value"), Text(data.UserEmail)),
				),
			),
			If(data.Paid, BillingCard(data)),
			If(!data.Paid, PaywallCard()),
		),
	)
}

func BillingCard(data SettingsData) Node {
	return Section(Class("settings-card"),
		H2(Class("settings-card-title"), Text("Billing")),
		Div(Class("settings-row"),
			Span(Text("Plan")),
			Span(Class("settings-value"), Text(planLabel(data.Paid))),
		),
		Div(Class("settings-row"),
			Span(Text("Usage balance")),
			Span(Class("settings-value"), Text(data.BalanceCents.Format())),
		),
		A(Class("settings-action"), Href("/settings/billing"), Text("Add usage credit")),
	)
}

func PaywallCard() Node {
	return Section(Class("settings-card"),
		H2(Class("settings-card-title"), Text("Unlock A Fast Chat")),
		P(Class("settings-copy"),
			Text(fmt.Sprintf(
				"Lifetime access is %s and includes %s of usage credit.",
				polar.ProductProPriceCents.Format(), polar.ProductProIncludedCreditCents.Format(),
			)),
		),
		Form(Method("post"), Action("/settings/checkout/pro"),
			Button(Type("submit"), Class("settings-action"), Text("Buy lifetime access")),
		),
	)
}

func CheckoutThankYouPage(confirmed bool) Node {
	return Main(Class("settings-page"),
		Div(Class("settings-shell"),
			Section(Class("settings-card"),
				H1(Class("settings-title"), Text(thankYouTitle(confirmed))),
				P(Class("settings-copy"), Text(thankYouCopy(confirmed))),
				A(Class("settings-action"), Href("/"), Text("Start chatting")),
			),
		),
	)
}

func planLabel(paid bool) string {
	if paid {
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

func thankYouCopy(confirmed bool) string {
	if confirmed {
		return fmt.Sprintf(
			"Your lifetime access is active and %s of usage credit was added to your account.",
			polar.ProductProIncludedCreditCents.Format(),
		)
	}
	return "We could not confirm the payment yet. Refresh this page in a moment."
}
