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
				Class("flex flex-col gap-4 rounded-xl border border-[color-mix(in_oklch,#262626_80%,#737373)] bg-[color-mix(in_oklch,#0c0a09_70%,#0a0a0a)] p-7 shadow-[0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)]"),
				P(Class("text-xs font-semibold uppercase leading-none tracking-[0.12em] text-neutral-500"), Text("Welcome back")),
				H1(Class("text-4xl font-bold leading-none tracking-[-0.04em] text-neutral-100"), Text("Log in")),
				P(
					Class("max-w-[21rem] text-base leading-7 text-neutral-400"),
					Text("Use your email and password to open afastchat."),
				),
				If(data.Error != "",
					P(Class("rounded-lg border border-red-700/30 bg-red-950/35 px-3 py-2 text-red-300"), Text(data.Error)),
				),
				Form(
					Class("flex flex-col gap-4"),
					Method("post"),
					Action("/login"),
					Div(Class("grid gap-2"),
						Label(Class("font-semibold text-neutral-400"), For("email"), Text("Email")),
						Input(
							Class("min-h-12 w-full rounded-lg border border-neutral-800 bg-[color-mix(in_oklch,#171717_88%,#0a0a0a)] px-3 font-[inherit] text-base text-neutral-100 outline-none placeholder:text-neutral-500 focus:border-[color-mix(in_oklch,#f97316_70%,white)] focus:shadow-[0_0_0_3px_color-mix(in_oklch,#f97316_22%,transparent)]"),
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
						Label(Class("font-semibold text-neutral-400"), For("password"), Text("Password")),
						Input(
							Class("min-h-12 w-full rounded-lg border border-neutral-800 bg-[color-mix(in_oklch,#171717_88%,#0a0a0a)] px-3 font-[inherit] text-base text-neutral-100 outline-none placeholder:text-neutral-500 focus:border-[color-mix(in_oklch,#f97316_70%,white)] focus:shadow-[0_0_0_3px_color-mix(in_oklch,#f97316_22%,transparent)]"),
							Type("password"),
							ID("password"),
							Name("password"),
							Placeholder("Password"),
							Required(),
						),
					),
					Button(Type("submit"), Class("inline-flex min-h-12 w-full cursor-pointer items-center justify-center rounded-lg border border-[color-mix(in_oklch,#f97316_80%,white)] bg-[linear-gradient(180deg,color-mix(in_oklch,#f97316_92%,white),#f97316)] px-5 font-[inherit] text-sm font-bold text-white no-underline transition hover:-translate-y-px hover:brightness-110"), Text("Log in")),
				),
				P(
					Class("text-center text-neutral-400"),
					Text("Need an account? "),
					A(Class("font-semibold text-neutral-100"), Href("/signup"), Text("Sign up")),
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
				Class("flex flex-col gap-4 rounded-xl border border-[color-mix(in_oklch,#262626_80%,#737373)] bg-[color-mix(in_oklch,#0c0a09_70%,#0a0a0a)] p-7 shadow-[0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)]"),
				P(Class("text-xs font-semibold uppercase leading-none tracking-[0.12em] text-neutral-500"), Text("Create account")),
				H1(Class("text-4xl font-bold leading-none tracking-[-0.04em] text-neutral-100"), Text("Sign up")),
				P(
					Class("max-w-[21rem] text-base leading-7 text-neutral-400"),
					Text("Create your account with email and password."),
				),
				If(data.Error != "",
					P(Class("rounded-lg border border-red-700/30 bg-red-950/35 px-3 py-2 text-red-300"), Text(data.Error)),
				),
				Form(
					Class("flex flex-col gap-4"),
					Method("post"),
					Action("/signup"),
					Div(Class("grid gap-2"),
						Label(Class("font-semibold text-neutral-400"), For("email"), Text("Email")),
						Input(
							Class("min-h-12 w-full rounded-lg border border-neutral-800 bg-[color-mix(in_oklch,#171717_88%,#0a0a0a)] px-3 font-[inherit] text-base text-neutral-100 outline-none placeholder:text-neutral-500 focus:border-[color-mix(in_oklch,#f97316_70%,white)] focus:shadow-[0_0_0_3px_color-mix(in_oklch,#f97316_22%,transparent)]"),
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
						Label(Class("font-semibold text-neutral-400"), For("password"), Text("Password")),
						Input(
							Class("min-h-12 w-full rounded-lg border border-neutral-800 bg-[color-mix(in_oklch,#171717_88%,#0a0a0a)] px-3 font-[inherit] text-base text-neutral-100 outline-none placeholder:text-neutral-500 focus:border-[color-mix(in_oklch,#f97316_70%,white)] focus:shadow-[0_0_0_3px_color-mix(in_oklch,#f97316_22%,transparent)]"),
							Type("password"),
							ID("password"),
							Name("password"),
							Placeholder("At least 8 characters"),
							Required(),
						),
					),
					Button(Type("submit"), Class("inline-flex min-h-12 w-full cursor-pointer items-center justify-center rounded-lg border border-[color-mix(in_oklch,#f97316_80%,white)] bg-[linear-gradient(180deg,color-mix(in_oklch,#f97316_92%,white),#f97316)] px-5 font-[inherit] text-sm font-bold text-white no-underline transition hover:-translate-y-px hover:brightness-110"), Text("Sign up")),
				),
				P(
					Class("text-center text-neutral-400"),
					Text("Already have an account? "),
					A(Class("font-semibold text-neutral-100"), Href("/login"), Text("Log in")),
				),
			),
		),
	)
}
