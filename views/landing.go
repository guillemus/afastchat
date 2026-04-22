package views

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func LandingPage() Node {
	return Main(
		Class("guest-main"),
		Div(
			Class("landing-shell"),
			Div(
				Class("landing-hero"),
				P(
					Class("landing-eyebrow"),
					Text("Fast local chat for your agents and threads"),
				),
				H1(
					Class("landing-title"),
					Text("afastchat"),
					Span(Class("logo-accent"), Text(".com")),
				),
				P(
					Class("landing-copy"),
					Text("Chat UI for long-running threads, streaming replies, and tool output."),
				),
				Div(
					Class("landing-actions"),
					A(Href("/login"), Class("btn-primary"), Text("Log in")),
					A(Href("/app"), Class("btn-secondary"), Text("Open app")),
				),
			),
		),
	)
}
