package views

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func MarkdownContent(mdContent string) Node {
	return Div(
		Class("markdown-content"),
		Raw(mdContent),
	)
}
