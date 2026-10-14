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
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

const copyCodeExpr = `
	navigator.clipboard.writeText(el.closest('.code-block').querySelector('pre').textContent);
	el.textContent = 'Copied';
	setTimeout(() => el.textContent = 'Copy', 1500);
`

const showHTMLPreviewExpr = `el.closest('html-preview').dataset.view = 'preview'`

const showHTMLCodeExpr = `el.closest('html-preview').dataset.view = 'code'`

var (
	md          = newMarkdown(true)
	mdStreaming = newMarkdown(false)
)

func newMarkdown(htmlPreview bool) goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			&katex.Extender{},
			highlighting.NewHighlighting(
				highlighting.WithFormatOptions(
					chromahtml.WithClasses(true),
				),
				highlighting.WithWrapperRenderer(codeBlockWrapper(htmlPreview)),
			),
		),
		goldmark.WithParserOptions(parser.WithInlineParsers(util.Prioritized(&inlineMathParser{}, -1))),
	)
}

type inlineMathParser struct {
	katex.Parser
}

func (p *inlineMathParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, segment := block.PeekLine()
	if (len(line) > 1 && line[1] == '$') || isInlineMath(line) {
		return p.Parser.Parse(parent, block, pc)
	}
	block.Advance(1)
	return ast.NewTextSegment(text.NewSegment(segment.Start, segment.Start+1))
}

func isInlineMath(line []byte) bool {
	if len(line) < 3 || util.IsSpace(line[1]) {
		return false
	}
	for i := 1; i < len(line); i++ {
		switch line[i] {
		case '\\':
			i++
		case '$':
			return !util.IsSpace(line[i-1]) && (i+1 >= len(line) || line[i+1] < '0' || line[i+1] > '9')
		}
	}
	return false
}

func codeBlockWrapper(htmlPreview bool) highlighting.WrapperRenderer {
	return func(w util.BufWriter, c highlighting.CodeBlockContext, entering bool) {
		label := "code"
		if language, ok := c.Language(); ok && len(language) > 0 {
			label = string(language)
		}
		preview := htmlPreview && strings.EqualFold(label, "html")

		if !entering {
			if !c.Highlighted() {
				_, _ = w.WriteString("</code></pre>")
			}
			if preview {
				_, _ = w.WriteString("</div></html-preview>\n")
			} else {
				_, _ = w.WriteString("</div>\n")
			}
			return
		}

		if preview {
			_, _ = w.WriteString(`<html-preview class="code-block" data-view="preview" data-ignore-morph>`)
		} else {
			_, _ = w.WriteString(`<div class="code-block">`)
		}
		_, _ = w.WriteString(`<div class="code-block-header"><span>`)
		_, _ = w.WriteString(html.EscapeString(label))
		_, _ = w.WriteString(`</span>`)
		if preview {
			writeCodeTab(w, "preview", "Preview", showHTMLPreviewExpr)
			writeCodeTab(w, "code", "Code", showHTMLCodeExpr)
		}
		_, _ = w.WriteString(`<button type="button" class="code-copy" data-on:click="`)
		_, _ = w.WriteString(html.EscapeString(copyCodeExpr))
		_, _ = w.WriteString(`">Copy</button></div>`)
		if preview {
			_, _ = w.WriteString(`<div class="html-preview-frame"></div><div class="html-preview-code">`)
		}
		if !c.Highlighted() {
			_, _ = w.WriteString("<pre><code>")
		}
	}
}

func writeCodeTab(w util.BufWriter, view, label, expr string) {
	_, _ = w.WriteString(`<button type="button" class="code-tab" data-tab="`)
	_, _ = w.WriteString(view)
	_, _ = w.WriteString(`" data-on:click="`)
	_, _ = w.WriteString(html.EscapeString(expr))
	_, _ = w.WriteString(`">`)
	_, _ = w.WriteString(label)
	_, _ = w.WriteString(`</button>`)
}

var tableWrapper = strings.NewReplacer(
	"<table>", `<div class="table-wrap"><table>`,
	"</table>", "</table></div>",
)

func renderMarkdown(src string) string {
	return convertMarkdown(md, src)
}

func renderStreamingMarkdown(src string) string {
	return convertMarkdown(mdStreaming, src)
}

func convertMarkdown(md goldmark.Markdown, src string) string {
	var buf bytes.Buffer
	b := unsafe.Slice(unsafe.StringData(src), len(src))
	if err := md.Convert(b, &buf); err != nil {
		slog.Error("error rendering markdown", "err", err)
		return html.EscapeString(src)
	}
	return tableWrapper.Replace(buf.String())
}
