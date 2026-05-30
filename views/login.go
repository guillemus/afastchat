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
		Class("grid min-h-dvh place-items-center p-4"),
		Div(
			Class("w-full max-w-[25rem]"),
			Div(
				Class("flex flex-col gap-4 rounded-xl border border-[color-mix(in_oklch,var(--border-color)_80%,var(--text-muted))] bg-[color-mix(in_oklch,var(--bg-sidebar)_70%,var(--bg-primary))] p-7 shadow-[var(--shadow-3)]"),
				P(Class("text-xs font-semibold uppercase leading-none tracking-[0.12em] text-[var(--text-muted)]"), Text("Welcome back")),
				H1(Class("text-4xl font-bold leading-none tracking-[-0.04em] text-[var(--text-primary)]"), Text("Log in")),
				P(
					Class("max-w-[21rem] text-base leading-7 text-[var(--text-secondary)]"),
					Text("Use your email and password to open afastchat."),
				),
				If(data.Error != "",
					P(Class("rounded-lg border border-[color-mix(in_oklch,var(--red-7)_30%,transparent)] bg-[color-mix(in_oklch,var(--red-9)_35%,transparent)] px-3 py-2 text-[var(--red-3)]"), Text(data.Error)),
				),
				Form(
					Class("flex flex-col gap-4"),
					Method("post"),
					Action("/login"),
					Div(Class("grid gap-2"),
						Label(Class("font-semibold text-[var(--text-secondary)]"), For("email"), Text("Email")),
						Input(
							Class("min-h-12 w-full rounded-lg border border-[var(--border-color)] bg-[color-mix(in_oklch,var(--bg-input)_88%,var(--bg-primary))] px-3 font-[inherit] text-base text-[var(--text-primary)] outline-none placeholder:text-[var(--text-muted)] focus:border-[color-mix(in_oklch,var(--bg-button)_70%,white)] focus:shadow-[0_0_0_3px_color-mix(in_oklch,var(--bg-button)_22%,transparent)]"),
							Type("email"),
							ID("email"),
							Name("email"),
							Placeholder("you@example.com"),
							Value(data.Email),
							Required(),
							AutoFocus(),
						),
					),
					Div(Class("grid gap-2"),
						Label(Class("font-semibold text-[var(--text-secondary)]"), For("password"), Text("Password")),
						Input(
							Class("min-h-12 w-full rounded-lg border border-[var(--border-color)] bg-[color-mix(in_oklch,var(--bg-input)_88%,var(--bg-primary))] px-3 font-[inherit] text-base text-[var(--text-primary)] outline-none placeholder:text-[var(--text-muted)] focus:border-[color-mix(in_oklch,var(--bg-button)_70%,white)] focus:shadow-[0_0_0_3px_color-mix(in_oklch,var(--bg-button)_22%,transparent)]"),
							Type("password"),
							ID("password"),
							Name("password"),
							Placeholder("Password"),
							Required(),
						),
					),
					Button(Type("submit"), Class("inline-flex min-h-12 w-full cursor-pointer items-center justify-center rounded-lg border border-[color-mix(in_oklch,var(--bg-button)_80%,white)] bg-[linear-gradient(180deg,color-mix(in_oklch,var(--bg-button)_92%,white),var(--bg-button))] px-5 font-[inherit] text-sm font-bold text-white no-underline transition hover:-translate-y-px hover:brightness-110"), Text("Log in")),
				),
				P(
					Class("text-center text-[var(--text-secondary)]"),
					Text("Need an account? "),
					A(Class("font-semibold text-[var(--text-primary)]"), Href("/signup"), Text("Sign up")),
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
		Class("grid min-h-dvh place-items-center p-4"),
		Div(
			Class("w-full max-w-[25rem]"),
			Div(
				Class("flex flex-col gap-4 rounded-xl border border-[color-mix(in_oklch,var(--border-color)_80%,var(--text-muted))] bg-[color-mix(in_oklch,var(--bg-sidebar)_70%,var(--bg-primary))] p-7 shadow-[var(--shadow-3)]"),
				P(Class("text-xs font-semibold uppercase leading-none tracking-[0.12em] text-[var(--text-muted)]"), Text("Create account")),
				H1(Class("text-4xl font-bold leading-none tracking-[-0.04em] text-[var(--text-primary)]"), Text("Sign up")),
				P(
					Class("max-w-[21rem] text-base leading-7 text-[var(--text-secondary)]"),
					Text("Create your account with email and password."),
				),
				If(data.Error != "",
					P(Class("rounded-lg border border-[color-mix(in_oklch,var(--red-7)_30%,transparent)] bg-[color-mix(in_oklch,var(--red-9)_35%,transparent)] px-3 py-2 text-[var(--red-3)]"), Text(data.Error)),
				),
				Form(
					Class("flex flex-col gap-4"),
					Method("post"),
					Action("/signup"),
					Div(Class("grid gap-2"),
						Label(Class("font-semibold text-[var(--text-secondary)]"), For("email"), Text("Email")),
						Input(
							Class("min-h-12 w-full rounded-lg border border-[var(--border-color)] bg-[color-mix(in_oklch,var(--bg-input)_88%,var(--bg-primary))] px-3 font-[inherit] text-base text-[var(--text-primary)] outline-none placeholder:text-[var(--text-muted)] focus:border-[color-mix(in_oklch,var(--bg-button)_70%,white)] focus:shadow-[0_0_0_3px_color-mix(in_oklch,var(--bg-button)_22%,transparent)]"),
							Type("email"),
							ID("email"),
							Name("email"),
							Placeholder("you@example.com"),
							Value(data.Email),
							Required(),
							AutoFocus(),
						),
					),
					Div(Class("grid gap-2"),
						Label(Class("font-semibold text-[var(--text-secondary)]"), For("password"), Text("Password")),
						Input(
							Class("min-h-12 w-full rounded-lg border border-[var(--border-color)] bg-[color-mix(in_oklch,var(--bg-input)_88%,var(--bg-primary))] px-3 font-[inherit] text-base text-[var(--text-primary)] outline-none placeholder:text-[var(--text-muted)] focus:border-[color-mix(in_oklch,var(--bg-button)_70%,white)] focus:shadow-[0_0_0_3px_color-mix(in_oklch,var(--bg-button)_22%,transparent)]"),
							Type("password"),
							ID("password"),
							Name("password"),
							Placeholder("At least 8 characters"),
							Required(),
						),
					),
					Button(Type("submit"), Class("inline-flex min-h-12 w-full cursor-pointer items-center justify-center rounded-lg border border-[color-mix(in_oklch,var(--bg-button)_80%,white)] bg-[linear-gradient(180deg,color-mix(in_oklch,var(--bg-button)_92%,white),var(--bg-button))] px-5 font-[inherit] text-sm font-bold text-white no-underline transition hover:-translate-y-px hover:brightness-110"), Text("Sign up")),
				),
				P(
					Class("text-center text-[var(--text-secondary)]"),
					Text("Already have an account? "),
					A(Class("font-semibold text-[var(--text-primary)]"), Href("/login"), Text("Log in")),
				),
			),
		),
	)
}
