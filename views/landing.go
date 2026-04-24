package views

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

var _ = styler.Style(`
	.landing-page {
		min-height: 100vh;
		min-height: 100dvh;
		display: grid;
		place-items: center;
		padding: var(--size-fluid-5);

		.landing-shell {
			width: min(100%, 48rem);
		}

		.landing-hero {
			display: flex;
			flex-direction: column;
			gap: var(--size-5);
			padding: var(--size-fluid-7);
			text-align: center;
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

		.landing-title {
			color: var(--text-primary);
			letter-spacing: -0.04em;
			line-height: 1;
			font-size: clamp(3rem, 10vw, 5.5rem);
		}

		.logo-accent {
			color: var(--bg-button);
			font-weight: var(--font-weight-4);
		}

		.landing-copy {
			color: var(--text-secondary);
			font-size: var(--font-size-1);
			line-height: var(--font-lineheight-3);
		}

		.landing-actions {
			display: flex;
			align-items: center;
			justify-content: center;
			gap: var(--size-3);
			flex-wrap: wrap;
		}

		.btn-primary,
		.btn-secondary {
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
			transition:
				filter 0.15s var(--ease-3),
				border-color 0.15s var(--ease-3),
				background-color 0.15s var(--ease-3);

			&:hover {
				filter: brightness(1.08);
			}
		}

		.btn-primary {
			background: var(--bg-button);
			color: white;
		}

		.btn-secondary {
			background: transparent;
			color: var(--text-primary);
			border-color: var(--border-color);
		}
	}
`)

func LandingPage() Node {
	return Main(
		Class("landing-page"),
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
					A(Href("/signup"), Class("btn-primary"), Text("Sign up")),
					A(Href("/login"), Class("btn-secondary"), Text("Log in")),
				),
			),
		),
	)
}
