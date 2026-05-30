package views

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func LandingPage() Node {
	pills := []string{
		"Explain a Go concurrency pattern",
		"Debug an error in my code",
		"Write a SQL migration",
		"Brainstorm system design ideas",
	}

	return Main(
		Class("flex min-h-dvh flex-col items-center justify-center bg-[var(--bg-primary)] p-6"),
		Div(
			Class("flex w-full max-w-3xl flex-col items-center gap-8"),
			Div(
				Class("lowercase tracking-[0.1em] text-[var(--text-secondary)]"),
				Text("afastchat"),
			),
			H1(
				Class("text-center text-4xl font-medium leading-tight tracking-[-0.04em]"),
				Text("What can I help with?"),
			),
			Div(
				Class("w-full max-w-3xl"),
				A(
					Href("/signup"),
					Class("flex items-center rounded-xl border border-[var(--border-color)] bg-[var(--bg-input)] p-3 text-[var(--text-muted)] no-underline transition-colors hover:border-[color-mix(in_oklch,var(--bg-button)_50%,var(--border-color))]"),
					Text("Message afastchat"),
					iconArrowUp("ml-auto h-5 w-5 shrink-0 text-[var(--text-muted)]"),
				),
			),
			Div(
				Class("flex w-full flex-wrap justify-center gap-3"),
				Map(pills, func(label string) Node {
					return A(Href("/signup"), Class("inline-flex items-center rounded-2xl border border-[var(--border-color)] px-4 py-2 text-[var(--text-secondary)] no-underline transition-colors hover:border-[var(--text-muted)] hover:bg-[var(--bg-hover)] hover:text-[var(--text-primary)]"), Text(label))
				}),
			),
			Div(
				Class("mt-2 flex items-center gap-5"),
				A(Href("/login"), Class("text-[var(--text-secondary)] no-underline transition-colors hover:text-[var(--text-primary)] hover:underline hover:underline-offset-3"), Text("Log in")),
				A(Href("/signup"), Class("text-[var(--text-secondary)] no-underline transition-colors hover:text-[var(--text-primary)] hover:underline hover:underline-offset-3"), Text("Sign up")),
			),
		),
	)
}
