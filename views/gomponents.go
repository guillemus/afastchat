package views

import (
	"afastchat.com/env"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func DebugSignals() Node {
	if !env.Env.Dev {
		return nil
	}

	return Div(
		Style(`
			background: black;
			position: absolute;	
			top: 0;
			right: 0;
			width: 320px;
			overflow: hidden;
		`),
		Small(Text("signals:")),
		Pre(
			Style(`
				margin: 0;
				white-space: pre;
				overflow: hidden;
			`),
			Data("json-signals", ""),
		),
	)
}
