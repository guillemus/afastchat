package views

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

var _ = styler.Style(`
	.auth-page {
		min-height: 100vh;
		min-height: 100dvh;
		display: grid;
		place-items: center;
		padding: var(--size-fluid-5);

		.auth-shell {
			width: min(100%, 48rem);
		}

		.auth-card {
			display: flex;
			flex-direction: column;
			gap: var(--size-4);
			padding: var(--size-fluid-6);
			width: min(100%, 30rem);
			margin: 0 auto;
			background: color-mix(in oklch, var(--bg-sidebar) 75%, var(--bg-primary));
			border: 1px solid var(--border-color);
			border-radius: var(--radius-4);
			box-shadow: var(--shadow-4);
		}

		.landing-eyebrow {
			color: var(--text-secondary);
			font-size: var(--font-size-0);
			text-transform: uppercase;
			letter-spacing: 0.08em;
		}

		.auth-title {
			color: var(--text-primary);
			letter-spacing: -0.04em;
			line-height: 1;
		}

		.auth-copy {
			color: var(--text-secondary);
			font-size: var(--font-size-1);
			line-height: var(--font-lineheight-3);
		}

		.auth-form {
			display: flex;
			flex-direction: column;
			gap: var(--size-2);

			label {
				color: var(--text-secondary);
				font-size: var(--font-size-0);
			}

			input {
				width: 100%;
				min-height: 44px;
				background: var(--bg-input);
				color: var(--text-primary);
				border: 1px solid var(--border-color);
				border-radius: var(--radius-2);
				padding: 0 var(--size-3);
				font-family: inherit;
				font-size: var(--font-size-0);
				outline: none;

				&::placeholder {
					color: var(--text-muted);
				}

				&:focus {
					border-color: color-mix(in oklch, var(--bg-button) 60%, var(--border-color));
					box-shadow: 0 0 0 3px color-mix(in oklch, var(--bg-button) 20%, transparent);
				}
			}
		}

		.auth-error {
			color: var(--red-3);
			background: color-mix(in oklch, var(--red-9) 35%, transparent);
			border: 1px solid color-mix(in oklch, var(--red-7) 30%, transparent);
			border-radius: var(--radius-2);
			padding: var(--size-2) var(--size-3);
			font-size: var(--font-size-0);
		}

		.btn-primary {
			display: inline-flex;
			align-items: center;
			justify-content: center;
			min-height: 44px;
			padding: 0 var(--size-5);
			border-radius: var(--radius-3);
			border: 1px solid transparent;
			font-family: inherit;
			font-size: var(--font-size-0);
			font-weight: var(--font-weight-6);
			text-decoration: none;
			cursor: pointer;
			transition: filter 0.15s var(--ease-3);
			background: var(--bg-button);
			color: white;

			&:hover {
				filter: brightness(1.08);
			}
		}

		.auth-submit {
			margin-top: var(--size-2);
			width: 100%;
		}

		.auth-footer {
			color: var(--text-secondary);
			font-size: var(--font-size-0);
			text-align: center;

			a {
				color: var(--text-primary);
			}
		}
	}
`)

type LoginPageData struct {
	Email string
	Error string
}

func LoginPage(data LoginPageData) Node {
	return Main(
		Class("auth-page"),
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

type SignupPageData struct {
	Email string
	Error string
}

func SignupPage(data SignupPageData) Node {
	return Main(
		Class("auth-page"),
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
