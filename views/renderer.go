package views

import (
	"io"
	"log/slog"
	"net/url"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

type Renderer struct {
	assetVersion string
}

func NewRenderer(assetVersion string) *Renderer {
	return &Renderer{
		assetVersion: url.QueryEscape(assetVersion),
	}
}

type (
	Thread struct {
		ID          string
		IsStreaming bool
		Title       string
	}
	Message struct {
		Type string
		ID   string
		// markdown rendered content
		MdContent string
		ToolName  string
	}
)

type IndexData struct {
	Threads []Thread
}

type ThreadData struct {
	Threads   []Thread
	Thread    Thread
	Messages  []Message
	Streaming bool
}

func (x *Renderer) RenderIndex(w io.Writer, data IndexData, doScrollSidebar bool) {
	renderNode(w, x.appLayout(
		sidebar(data.Threads, "", doScrollSidebar),
		threadContentsEmpty(),
	))
}

func (x *Renderer) RenderThread(w io.Writer, data ThreadData, doScrollSidebar bool) {
	renderNode(w, x.appLayout(
		sidebar(data.Threads, data.Thread.ID, doScrollSidebar),
		threadContents(data.Messages, data.Thread.ID, doScrollSidebar, data.Streaming),
	))
}

func renderNode(w io.Writer, node Node) {
	err := node.Render(w)
	if err != nil {
		slog.Error("view render error", "err", err)
	}
}

func RenderThreads(w io.Writer, threads []Thread, activeThreadID string) {
	for _, it := range threads {
		renderNode(w, threadListItem(it, it.ID == activeThreadID))
	}
}

func RenderThreadItem(w io.Writer, thread Thread) {
	renderNode(w, threadListItem(thread, true))
}

func (x *Renderer) appLayout(body ...Node) Node {
	return HTML5(HTML5Props{
		Title:       "afastchat",
		Description: "",
		Language:    "en",
		Head: []Node{
			Link(Rel("preconnect"), Href("https://fonts.googleapis.com")),
			Link(Rel("preconnect"), Href("https://fonts.gstatic.com"), Attr("crossorigin")),
			Link(
				Rel("stylesheet"),
				Href("https://fonts.googleapis.com/css2?family=Fira+Code:wght@300..700&display=swap"),
			),
			Link(Rel("stylesheet"), Href("https://unpkg.com/open-props@1.7.23/open-props.min.css")),
			Link(Rel("stylesheet"), Href("/public/styles.css?v="+x.assetVersion)),
			Script(Type("module"), Src("/public/main.js?v="+x.assetVersion)),
			Script(
				Type("module"),
				Src("https://cdn.jsdelivr.net/gh/starfederation/datastar@1.0.0-RC.8/bundles/datastar.js"),
			),
		},
		Body: []Node{
			DebugSignals(),
			Div(Class("app"), Group(body)),
		},
	})
}

func RenderThreadMessages(w io.Writer, messages []Message, doScrollSidebar bool) {
	renderNode(w, ThreadMessages(messages, doScrollSidebar))
}
