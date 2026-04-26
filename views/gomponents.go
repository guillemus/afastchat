package views

import (
	"strings"

	"afastchat.com/env"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type Styles struct {
	styleString strings.Builder
}

var styles Styles

func (x *Styles) Style(css string) uint8 {
	x.styleString.WriteString(css)
	return 0
}

func (x *Styles) Node() Node {
	return StyleEl(Raw(x.styleString.String()))
}

var _ = styles.Style(`
	.debug-signals {
		background: var(--gray-12);
		position: absolute;
		top: 0;
		right: 0;
		width: 320px;
		padding: var(--size-2);
		overflow: hidden;
		z-index: 10;

		.debug-signals-header {
			display: flex;
			align-items: center;
			justify-content: space-between;
			gap: var(--size-2);
		}

		.debug-signals-toggle {
			background: transparent;
			border: 1px solid var(--border-color);
			border-radius: var(--radius-1);
			color: var(--text-primary);
			padding: 0 var(--size-2);
			cursor: pointer;
			font: inherit;
		}

		.signals-text {
			background: var(--gray-12);
			margin: 0;
			white-space: pre;
			overflow: hidden;
		}
	}
`)

func DebugSignals(env env.Env) Node {
	if !env.Dev {
		return nil
	}

	return Div(
		Class("debug-signals"),
		Data("signals:debug-signals-collapsed", "false"),
		Data("init", `$debugSignalsCollapsed = getSessionVal('debug_signals_collapsed') === 'true'`),
		Div(
			Class("debug-signals-header"),
			Small(Text("signals:")),
			Button(
				Type("button"),
				Class("debug-signals-toggle"),
				Data("text", "$debugSignalsCollapsed ? 'show' : 'hide'"),
				Data("on:click", `
					$debugSignalsCollapsed = !$debugSignalsCollapsed;
					setSessionVal('debug_signals_collapsed', $debugSignalsCollapsed ? 'true' : 'false');
				`),
			),
		),
		Pre(
			Class("signals-text"),
			Data("show", "!$debugSignalsCollapsed"),
			Data("json-signals", ""),
		),
	)
}
