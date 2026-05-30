package views

import (
	"io"
	"strings"
	"sync"

	. "maragu.dev/gomponents"
)

// Static pre-renders a large static HTML tree once and reuses the rendered HTML.
// Use it only for nodes that never depend on request data, user data, or changing state.
func Static(node Node) Node {
	var once sync.Once
	var html string
	var renderErr error

	return NodeFunc(func(w io.Writer) error {
		once.Do(func() {
			var b strings.Builder
			renderErr = node.Render(&b)
			html = b.String()
		})
		if renderErr != nil {
			return renderErr
		}
		_, writeErr := io.WriteString(w, html)
		return writeErr
	})
}
