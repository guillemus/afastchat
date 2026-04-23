package views

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type LoginPageData struct {
	Email string
	Error string
}

func LoginPage(data LoginPageData) Node {
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
					Text("Use your email and password to open afastchat."),
				),
				If(data.Error != "",
					P(Class("auth-error"), Text(data.Error)),
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
						Value(data.Email),
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
					Text("Need an account? "),
					A(Href("/signup"), Text("Sign up")),
				),
			),
		),
	)
}
