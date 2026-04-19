package server

import (
	"bytes"
	"html"
	"log/slog"
	"unsafe"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
)

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		highlighting.NewHighlighting(
			highlighting.WithFormatOptions(
				chromahtml.WithClasses(true),
			),
		),
	),
)

func renderMarkdown(src string) string {
	var buf bytes.Buffer
	b := unsafe.Slice(unsafe.StringData(src), len(src))
	if err := md.Convert(b, &buf); err != nil {
		slog.Error("error rendering markdown", "err", err)
		return html.EscapeString(src)
	}
	return buf.String()
}
