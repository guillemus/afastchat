package server

import (
	"bytes"
	"html"
	"log/slog"
	"strings"
	"unsafe"

	katex "github.com/FurqanSoftware/goldmark-katex"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/util"
)

const copyCodeExpr = `
	navigator.clipboard.writeText(el.closest('.code-block').querySelector('pre').textContent);
	el.textContent = 'Copied';
	setTimeout(() => el.textContent = 'Copy', 1500);
`

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		&katex.Extender{},
		highlighting.NewHighlighting(
			highlighting.WithFormatOptions(
				chromahtml.WithClasses(true),
			),
			highlighting.WithWrapperRenderer(renderCodeBlockWrapper),
		),
	),
)

func renderCodeBlockWrapper(w util.BufWriter, c highlighting.CodeBlockContext, entering bool) {
	if !entering {
		if !c.Highlighted() {
			_, _ = w.WriteString("</code></pre>")
		}
		_, _ = w.WriteString("</div>\n")
		return
	}

	label := "code"
	if language, ok := c.Language(); ok && len(language) > 0 {
		label = string(language)
	}

	_, _ = w.WriteString(`<div class="code-block"><div class="code-block-header"><span>`)
	_, _ = w.WriteString(html.EscapeString(label))
	_, _ = w.WriteString(`</span><button type="button" class="code-copy" data-on:click="`)
	_, _ = w.WriteString(html.EscapeString(copyCodeExpr))
	_, _ = w.WriteString(`">Copy</button></div>`)
	if !c.Highlighted() {
		_, _ = w.WriteString("<pre><code>")
	}
}

var tableWrapper = strings.NewReplacer(
	"<table>", `<div class="table-wrap"><table>`,
	"</table>", "</table></div>",
)

func renderMarkdown(src string) string {
	var buf bytes.Buffer
	b := unsafe.Slice(unsafe.StringData(src), len(src))
	if err := md.Convert(b, &buf); err != nil {
		slog.Error("error rendering markdown", "err", err)
		return html.EscapeString(src)
	}
	return tableWrapper.Replace(buf.String())
}
