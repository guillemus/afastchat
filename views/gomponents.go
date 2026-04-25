package views

import (
	"strings"

	"afastchat.com/env"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type Styler struct {
	styleString strings.Builder
}

var styler Styler

func (x *Styler) Style(css string) uint8 {
	x.styleString.WriteString(css)
	return 0
}

func (x *Styler) Node() Node {
	return StyleEl(Raw(x.styleString.String()))
}

var _ = styler.Style(`
	.debug-signals {
		background: var(--gray-12);
		position: absolute;
		top: 0;
		right: 0;
		width: 320px;
		padding: var(--size-2);
		overflow: hidden;
		z-index: 10;

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
		Small(Text("signals:")),
		Pre(
			Class("signals-text"),
			Data("json-signals", ""),
		),
	)
}
