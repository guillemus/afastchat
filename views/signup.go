package views

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type SignupPageData struct {
	Email string
	Error string
}

func SignupPage(data SignupPageData) Node {
	return Main(
		Class("guest-main"),
		Div(
			Class("auth-shell"),
			Div(
				Class("auth-card"),
				P(Class("landing-eyebrow"), Text("Create account")),
				H1(Class("auth-title"), Text("Sign up")),
				P(
					Class("auth-copy"),
					Text("Create your account with email and password."),
				),
				If(data.Error != "",
					P(Class("auth-error"), Text(data.Error)),
				),
				Form(
					Class("auth-form"),
					Method("post"),
					Action("/signup"),
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
						Placeholder("At least 8 characters"),
						Required(),
					),
					Button(Type("submit"), Class("btn-primary auth-submit"), Text("Sign up")),
				),
				P(
					Class("auth-footer"),
					Text("Already have an account? "),
					A(Href("/login"), Text("Log in")),
				),
			),
		),
	)
}
