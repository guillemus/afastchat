package views

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

var _ = styles.Style(`
	.landing-page {
		min-height: 100vh;
		min-height: 100dvh;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--size-6);
		background: var(--bg-primary);

		.landing-shell {
			width: min(100%, 768px);
			display: flex;
			flex-direction: column;
			align-items: center;
			gap: var(--size-8);
		}

		.landing-brand {
			color: var(--text-secondary);
			font-size: var(--font-size-0);
			letter-spacing: 0.1em;
			text-transform: lowercase;
		}

		.landing-title {
			color: var(--text-primary);
			font-size: clamp(2rem, 5vw, 2.75rem);
			font-weight: var(--font-weight-5);
			letter-spacing: -0.04em;
			line-height: 1.1;
			text-align: center;
		}

		.landing-input-wrap {
			width: 100%;
			max-width: 768px;
		}

		.landing-input {
			display: flex;
			align-items: center;
			background: var(--bg-input);
			border: 1px solid var(--border-color);
			border-radius: var(--radius-3);
			padding: var(--size-3);
			color: var(--text-muted);
			font-size: var(--font-size-1);
			cursor: pointer;
			text-decoration: none;
			transition: border-color 0.2s var(--ease-3);

			&:hover {
				border-color: color-mix(in oklch, var(--bg-button) 50%, var(--border-color));
			}

			svg {
				width: 20px;
				height: 20px;
				margin-left: auto;
				color: var(--text-muted);
				flex-shrink: 0;
			}
		}

		.landing-pills {
			display: flex;
			flex-wrap: wrap;
			justify-content: center;
			gap: var(--size-3);
			width: 100%;
		}

		.landing-pill {
			display: inline-flex;
			align-items: center;
			padding: var(--size-2) var(--size-4);
			background: transparent;
			border: 1px solid var(--border-color);
			border-radius: var(--radius-4);
			color: var(--text-secondary);
			font-size: var(--font-size-0);
			text-decoration: none;
			transition:
				border-color 0.15s var(--ease-3),
				color 0.15s var(--ease-3),
				background-color 0.15s var(--ease-3);
			cursor: pointer;

			&:hover {
				background: var(--bg-hover);
				border-color: var(--text-muted);
				color: var(--text-primary);
			}
		}

		.landing-actions {
			display: flex;
			align-items: center;
			gap: var(--size-5);
			margin-top: var(--size-2);
		}

		.landing-action {
			color: var(--text-secondary);
			font-size: var(--font-size-0);
			text-decoration: none;
			transition: color 0.15s var(--ease-3);

			&:hover {
				color: var(--text-primary);
				text-decoration: underline;
				text-underline-offset: 3px;
			}
		}
	}
`)

func LandingPage() Node {
	pills := []string{
		"Explain a Go concurrency pattern",
		"Debug an error in my code",
		"Write a SQL migration",
		"Brainstorm system design ideas",
	}

	return Main(
		Class("landing-page"),
		Div(
			Class("landing-shell"),
			Div(
				Class("landing-brand"),
				Text("afastchat"),
			),
			H1(
				Class("landing-title"),
				Text("What can I help with?"),
			),
			Div(
				Class("landing-input-wrap"),
				A(
					Href("/signup"),
					Class("landing-input"),
					Text("Message afastchat"),
					iconArrowUp(),
				),
			),
			Div(
				Class("landing-pills"),
				Map(pills, func(label string) Node {
					return A(Href("/signup"), Class("landing-pill"), Text(label))
				}),
			),
			Div(
				Class("landing-actions"),
				A(Href("/login"), Class("landing-action"), Text("Log in")),
				A(Href("/signup"), Class("landing-action"), Text("Sign up")),
			),
		),
	)
}
