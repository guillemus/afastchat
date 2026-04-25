package views

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

var _ = styler.Style(`
	.markdown-content {
		pre {
			background: var(--gray-11);
			border-radius: var(--radius-2);
			padding: var(--size-3);
			max-width: 100%;
			overflow-x: auto;
			margin: var(--size-2) 0;
		}

		code {
			font-family: 'Fira Code', var(--font-mono);
			font-size: var(--font-size-0);
		}

		:not(pre) > code {
			background: var(--gray-10);
			padding: var(--size-1) var(--size-2);
			border-radius: var(--radius-1);
		}

		p {
			margin: var(--size-2) 0;
		}
		p:first-child {
			margin-top: 0;
		}
		p:last-child {
			margin-bottom: 0;
		}

		ul,
		ol {
			padding-left: var(--size-5);
			margin: var(--size-2) 0;
		}

		li {
			margin: var(--size-1) 0;
		}

		h1,
		h2,
		h3,
		h4 {
			margin: var(--size-3) 0 var(--size-2);
			color: var(--text-primary);
		}

		h1 {
			font-size: var(--font-size-4);
		}
		h2 {
			font-size: var(--font-size-3);
		}
		h3 {
			font-size: var(--font-size-2);
		}

		blockquote {
			border-left: 3px solid var(--gray-8);
			padding-left: var(--size-3);
			color: var(--text-secondary);
			margin: var(--size-2) 0;
		}

		a {
			color: var(--bg-button);
			text-decoration: underline;
		}

		table {
			border-collapse: collapse;
			margin: var(--size-2) 0;
			width: 100%;
		}

		th,
		td {
			border: 1px solid var(--border-color);
			padding: var(--size-1) var(--size-2);
			text-align: left;
		}

		th {
			background: var(--gray-11);
		}

		hr {
			border: none;
			border-top: 1px solid var(--border-color);
			margin: var(--size-3) 0;
		}

		.chroma {
			background: var(--gray-11);
			color: var(--gray-3);
		}

		.chroma .k,
		.chroma .kd,
		.chroma .kn,
		.chroma .kc,
		.chroma .kr {
			color: var(--bg-button);
			font-weight: var(--font-weight-6);
		}

		.chroma .s,
		.chroma .sa,
		.chroma .sb,
		.chroma .sc,
		.chroma .sd,
		.chroma .s2,
		.chroma .se,
		.chroma .sh,
		.chroma .si,
		.chroma .sx,
		.chroma .sr,
		.chroma .s1,
		.chroma .ss {
			color: #7ee787;
		}

		.chroma .nf,
		.chroma .fm {
			color: var(--gray-1);
		}

		.chroma .c,
		.chroma .ch,
		.chroma .cm,
		.chroma .c1,
		.chroma .cs,
		.chroma .cp,
		.chroma .cpf {
			color: var(--gray-7);
			font-style: italic;
		}

		.chroma .nc,
		.chroma .kt,
		.chroma .nn {
			color: #ff7b72;
		}

		.chroma .mi,
		.chroma .mf,
		.chroma .mh,
		.chroma .mo,
		.chroma .mb,
		.chroma .il,
		.chroma .m {
			color: #79c0ff;
		}

		.chroma .nt,
		.chroma .na {
			color: #d2a8ff;
		}

		.chroma .nb,
		.chroma .bp {
			color: var(--gray-5);
		}

		.chroma .nv,
		.chroma .vc,
		.chroma .vg,
		.chroma .vi,
		.chroma .vm {
			color: var(--gray-4);
		}

		.chroma .o,
		.chroma .ow,
		.chroma .p {
			color: var(--gray-5);
		}
	}
`)

func MarkdownContent(mdContent string) Node {
	return Div(
		Class("markdown-content"),
		Raw(mdContent),
	)
}
