package views

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func LoginPage() Node {
	return Main(
		Class("guest-main"),
		Div(
			Class("auth-shell"),
			Div(
				Class("auth-card"),
				P(Class("landing-eyebrow"), Text("Welcome back")),
				H1(Class("auth-title"), Text("Log in")),
				P(
					Class("auth-copy"),
					Text("Traditional email and password login. Session cookie next."),
				),
				Form(
					Class("auth-form"),
					Method("post"),
					Action("/login"),
					Label(For("email"), Text("Email")),
					Input(
						Type("email"),
						ID("email"),
						Name("email"),
						Placeholder("you@example.com"),
						Required(),
						AutoFocus(),
					),
					Label(For("password"), Text("Password")),
					Input(
						Type("password"),
						ID("password"),
						Name("password"),
						Placeholder("Password"),
						Required(),
					),
					Button(Type("submit"), Class("btn-primary auth-submit"), Text("Log in")),
				),
				P(
					Class("auth-footer"),
					A(Href("/"), Text("Back to landing")),
				),
			),
		),
	)
}
