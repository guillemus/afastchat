package views

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strconv"

	"datastar-chat/db"

	"github.com/google/uuid"
	g "maragu.dev/gomponents"
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

type IndexData struct {
	Threads []db.Thread
}

type ThreadData struct {
	Threads   []db.Thread
	Thread    db.Thread
	Messages  []db.Message
	Streaming bool
}

func handleErr(err error) {
	if err != nil {
		slog.Error("view render error", "err", err)
	}
}

func (r *Renderer) RenderIndex(w io.Writer, data IndexData) {
	err := r.indexPage(data).Render(w)
	handleErr(err)
}

func (r *Renderer) RenderThread(w io.Writer, data ThreadData) {
	err := r.threadPage(data).Render(w)
	handleErr(err)
}

func (r *Renderer) RenderThreadItem(thread db.Thread, activeThreadID uuid.UUID) string {
	return renderString(threadItem(thread, activeThreadID))
}

func RenderUserMessage(messageID uuid.UUID, content string, isNew bool) string {
	return renderString(userMessage(messageID, content, isNew))
}

func RenderAssistantMessage(messageID uuid.UUID, contentHTML string, isNew bool) string {
	return renderString(assistantMessage(messageID, contentHTML, isNew))
}

func renderString(node g.Node) string {
	var b bytes.Buffer
	handleErr(node.Render(&b))

	return b.String()
}

func (r *Renderer) indexPage(data IndexData) g.Node {
	return page(r.assetVersion,
		appShell(
			"",
			false,
			"",
			sidebar(data.Threads, uuid.Nil),
			Main(
				Class("main"),
				Div(
					ID("messages"),
					Class("main-body"),
					scrollAnchor(),
				),
				chatInput(),
			),
		),
	)
}

func (r *Renderer) threadPage(data ThreadData) g.Node {
	return page(r.assetVersion,
		appShell(
			fmt.Sprintf("'%s'", data.Thread.ID),
			data.Streaming,
			fmt.Sprintf("@get('/threads/%s/stream')", data.Thread.ID),
			sidebar(data.Threads, data.Thread.ID),
			Main(
				Class("main"),
				Div(
					ID("messages"),
					Class("main-body"),
					messageList(data.Messages),
					scrollAnchor(),
				),
				chatInput(),
			),
		),
	)
}

func appShell(threadID string, loading bool, initAction string, children ...g.Node) g.Node {
	nodes := []g.Node{
		Class("app"),
		g.Attr("data-signals:message", "''"),
		g.Attr("data-signals:threadId", threadID),
		g.Attr("data-signals:threadSearch", "''"),
		g.Attr("data-signals:loading", strconv.FormatBool(loading)),
	}
	if initAction != "" {
		nodes = append(nodes, g.Attr("data-init__delay.25ms", initAction))
	}
	nodes = append(nodes, children...)

	return Div(nodes...)
}

func sidebar(threads []db.Thread, activeThreadID uuid.UUID) g.Node {
	return Aside(
		Class("sidebar"),
		Div(
			Class("sidebar-header"),
			Span(
				Class("logo"),
				g.Text("datastar"),
				Span(Class("logo-accent"), g.Text(".chat")),
			),
		),
		A(
			Href("/"),
			Class("new-chat-btn"),
			iconPlus(),
			g.Text("New Chat"),
		),
		Div(
			Class("search-box"),
			iconSearch(),
			Input(
				Type("text"),
				Placeholder("Search threads..."),
				g.Attr("data-bind:threadSearch"),
				g.Attr("data-on:input__debounce.150ms", "@get('/threads/search')"),
			),
		),
		Div(
			Class("thread-list"),
			Div(
				ID("thread-section-recent"),
				Class("thread-section-label"),
				g.Text("Recent"),
			),
			Div(
				ID("thread-items"),
				g.Map(threads, func(thread db.Thread) g.Node {
					return threadItem(thread, activeThreadID)
				}),
			),
		),
		Div(
			Class("sidebar-footer"),
			Div(Class("avatar"), g.Text("G")),
			Span(Class("sidebar-footer-name"), g.Text("Guillem")),
		),
	)
}

func threadItem(thread db.Thread, activeThreadID uuid.UUID) g.Node {
	className := "thread-item"
	if thread.ID == activeThreadID {
		className += " active"
	}

	return A(
		ID("thread-"+thread.ID.String()),
		Href("/threads/"+thread.ID.String()),
		Class(className),
		Span(Class("thread-item-text"), g.Text(thread.Title)),
		Button(
			Type("button"),
			Class("archive-btn"),
			g.Attr("data-on:click", fmt.Sprintf("event.preventDefault(); event.stopPropagation(); @post('/threads/%s/archive')", thread.ID)),
			iconArchive(),
		),
	)
}

func chatInput() g.Node {
	return Div(
		Class("chat-input-area"),
		Div(
			Class("chat-input-box"),
			Textarea(
				Rows("1"),
				AutoFocus(),
				Placeholder("Type your message here..."),
				g.Attr("data-bind:message"),
				g.Attr("data-attr:disabled", "$loading"),
				g.Attr("data-on:keydown", `
					if (event.key !== 'Enter') return;
					if (event.shiftKey) return;
					if ($loading) return;
					event.preventDefault();
					@post('/chat/send');
					$message = '';
				`),
			),
		),
	)
}

func messageList(messages []db.Message) g.Node {
	return g.Map(messages, func(message db.Message) g.Node {
		switch message.Role {
		case "user":
			if message.Content == "" {
				return nil
			}
			return userMessage(message.ID, message.Content, false)
		case "assistant":
			return assistantMessage(message.ID, message.Content, false)
		default:
			return nil
		}
	})
}

func userMessage(messageID uuid.UUID, content string, isNew bool) g.Node {
	className := "message user-message"
	if isNew {
		className += " is-new"
	}

	return Div(
		ID("msg-"+messageID.String()),
		Class(className),
		g.Text(content),
	)
}

func assistantMessage(messageID uuid.UUID, contentHTML string, isNew bool) g.Node {
	className := "message assistant-message"
	if isNew {
		className += " is-new"
	}

	return Div(
		ID("msg-"+messageID.String()),
		Class(className),
		g.Raw(contentHTML),
	)
}

func scrollAnchor() g.Node {
	return Div(
		ID("scroll-anchor"),
		Class("scroll-anchor"),
	)
}

func iconPlus() g.Node {
	return g.El("svg",
		g.Attr("xmlns", "http://www.w3.org/2000/svg"),
		g.Attr("viewBox", "0 0 24 24"),
		g.Attr("fill", "none"),
		g.Attr("stroke", "currentColor"),
		g.Attr("stroke-width", "2.5"),
		g.Attr("stroke-linecap", "round"),
		g.El("line", g.Attr("x1", "12"), g.Attr("y1", "5"), g.Attr("x2", "12"), g.Attr("y2", "19")),
		g.El("line", g.Attr("x1", "5"), g.Attr("y1", "12"), g.Attr("x2", "19"), g.Attr("y2", "12")),
	)
}

func iconSearch() g.Node {
	return g.El("svg",
		g.Attr("xmlns", "http://www.w3.org/2000/svg"),
		g.Attr("viewBox", "0 0 24 24"),
		g.Attr("fill", "none"),
		g.Attr("stroke", "currentColor"),
		g.Attr("stroke-width", "2"),
		g.El("circle", g.Attr("cx", "11"), g.Attr("cy", "11"), g.Attr("r", "8")),
		g.El("path", g.Attr("d", "m21 21-4.3-4.3")),
	)
}

func iconArchive() g.Node {
	return g.El("svg",
		g.Attr("xmlns", "http://www.w3.org/2000/svg"),
		g.Attr("viewBox", "0 0 24 24"),
		g.Attr("fill", "none"),
		g.Attr("stroke", "currentColor"),
		g.Attr("stroke-width", "2"),
		g.Attr("stroke-linecap", "round"),
		g.Attr("stroke-linejoin", "round"),
		g.El("polyline", g.Attr("points", "3 6 5 6 21 6")),
		g.El("path", g.Attr("d", "M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6")),
		g.El("line", g.Attr("x1", "10"), g.Attr("y1", "11"), g.Attr("x2", "10"), g.Attr("y2", "17")),
		g.El("line", g.Attr("x1", "14"), g.Attr("y1", "11"), g.Attr("x2", "14"), g.Attr("y2", "17")),
	)
}
