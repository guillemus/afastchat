package views

import (
	_ "embed"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

//go:embed scripts.js
var scriptsJS string

var JS = Script(Raw(scriptsJS))
