package views

import (
	"afastchat.com/env"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func DebugSignals(env env.Env) Node {
	if !env.Dev {
		return nil
	}

	return Div(
		Class("fixed right-5 bottom-2 z-10 w-[500px] overflow-hidden bg-[var(--gray-12)] p-2"),
		Data("signals:debug-signals-collapsed", "false"),
		Data("init", `$debugSignalsCollapsed = getSessionVal('debug_signals_collapsed') !== 'false'`),
		Div(
			Class("flex items-center justify-between gap-2"),
			Small(Text("signals:")),
			Button(
				Type("button"),
				Class("cursor-pointer rounded border border-[var(--border-color)] bg-transparent px-2 font-[inherit] text-[var(--text-primary)]"),
				Data("text", "$debugSignalsCollapsed ? 'show' : 'hide'"),
				Data("on:click", `
					$debugSignalsCollapsed = !$debugSignalsCollapsed;
					setSessionVal('debug_signals_collapsed', $debugSignalsCollapsed ? 'true' : 'false');
				`),
			),
		),
		Pre(
			Class("overflow-hidden whitespace-pre bg-[var(--gray-12)]"),
			Data("show", "!$debugSignalsCollapsed"),
			Data("json-signals", ""),
		),
	)
}
