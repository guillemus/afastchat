package views

import (
	"fmt"
	"io"
	"log/slog"
	"net/url"

	"afastchat.com/db"
	"afastchat.com/env"

	. "maragu.dev/gomponents"
	"maragu.dev/gomponents/components"
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

func (x *Renderer) RenderIndex(w io.Writer, data IndexData, doScrollSidebar bool) {
	renderNode(w, x.page(
		sidebar(data.Threads, "", doScrollSidebar),
		threadContentsEmpty(),
	))
}

func (x *Renderer) RenderThread(w io.Writer, data ThreadData, doScrollSidebar bool) {
	renderNode(w, x.page(
		sidebar(data.Threads, data.Thread.ID, doScrollSidebar),
		threadContents(data.Messages, data.Thread.ID, doScrollSidebar),
	))
}

func renderNode(w io.Writer, node Node) {
	err := node.Render(w)
	if err != nil {
		slog.Error("view render error", "err", err)
	}
}

func RenderThreads(w io.Writer, threads []db.Thread, activeThreadID string) {
	for _, it := range threads {
		renderNode(w, threadListItem(it, it.ID == activeThreadID))
	}
}

func RenderThreadItem(w io.Writer, thread db.Thread) {
	renderNode(w, threadListItem(thread, false))
}

func (x *Renderer) page(body ...Node) Node {
	return components.HTML5(components.HTML5Props{
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
			Script(
				Type("module"),
				Src("https://cdn.jsdelivr.net/gh/starfederation/datastar@1.0.0-RC.8/bundles/datastar.js"),
			),
			Script(Type("module"), Src("/public/main.js?v="+x.assetVersion)),
		},
		Body: []Node{
			DebugSignals(),
			Div(Class("app"), Group(body)),
		},
	})
}

func sidebar(threads []db.Thread, activeThreadID string, doScrollSidebar bool) Node {
	return Aside(
		Class("sidebar"),
		If(activeThreadID != "",
			Data("init", fmt.Sprintf("@get('/threads/%s/events')", activeThreadID)),
		),

		Div(Class("sidebar-header"),
			Span(
				Class("logo"),
				Text("afastchat"),
				Span(Class("logo-accent"), Text(".com")),
			),
		),
		A(Href("/"), Class("new-chat-btn"),
			iconPlus(),
			Text("New Chat"),
		),
		Div(Class("search-box"),
			iconSearch(),
			Input(
				Type("text"),
				Name("thread-search"),
				Placeholder("Search threads..."),
				Data("bind:thread-search", ""),
				Data("on:input__debounce.150ms", "@get('/threads/search')"),
			),
		),
		Div(
			components.Classes{
				"thread-list": true,
				"invisible":   doScrollSidebar,
			},
			If(doScrollSidebar, Group{
				Data("init", `
					const value = getSessionVal('thread_list_scroll');
					if (value) {
						el.scrollTop = Number.parseInt(value, 10) || 0;
					}

					el.classList.remove('invisible');
				`),
			}),

			Data("on:scroll__debounce.10ms", `
				let val = String(Math.round(el.scrollTop));
				setSessionVal('thread_list_scroll', val);
			`),

			Div(
				ID("thread-section-recent"),
				Class("thread-section-label"),
				Text("Recent"),
			),
			Div(ID("thread-items"),
				Map(threads, func(thread db.Thread) Node {
					return threadListItem(thread, thread.ID == activeThreadID)
				}),
			),
		),
		Div(
			Class("sidebar-footer"),
			Div(Class("avatar"), Text("G")),
			Span(Class("sidebar-footer-name"), Text("Guillem")),
		),
	)
}

func threadListItem(thread db.Thread, isActive bool) Node {
	className := "thread-item"
	if isActive {
		className += " active"
	}

	return A(
		ID("thread-"+thread.ID),
		Class(className),
		Href(fmt.Sprintf("/threads/%s", thread.ID)),

		Span(Class("thread-item-text"), Text(thread.Title)),
		Button(
			Type("button"),
			Class("archive-btn"),
			Data("on:click", fmt.Sprintf(`
				evt.preventDefault();
				evt.stopPropagation();
				@post('/threads/%s/archive');
			`, thread.ID)),
			iconArchive(),
		),
	)
}

func threadContentsEmpty() Node {
	return threadContents(nil, "", false)
}

func threadContents(messages []db.Message, threadID string, doScrollSidebar bool) Node {
	var contents Node
	if len(messages) == 0 {
		contents = Group{
			Div(
				ID("messages"),
				Class("main-body empty-thread"),
				Div(
					Class("empty-state"),
					H1(Class("empty-state-title"), Text("How can I help?")),
					P(Class("empty-state-subtitle"),
						Text("Start a new conversation with afastchat."),
					),
				),
			),
			newThreadChatInput(),
		}
	} else {
		contents = Group{
			Div(
				ID("messages"),
				components.Classes{
					"main-body": true,
					"invisible": doScrollSidebar,
				},
				Class("main-body invisible"),
				If(doScrollSidebar, Data("init", `
					el.scrollTop = el.scrollHeight;
					el.classList.remove('invisible');
				`)),

				Map(messages, func(message db.Message) Node {
					switch message.Role {
					case "user":
						return userMessage(message.ID, message.Content)
					case "assistant":
						return assistantMessage(message.ID, message.Content, false)
					default:
						return nil
					}
				}),
			),
			existingThreadChatInput(threadID),
		}
	}

	return Main(Class("main"), contents)
}

type chatInputProps struct {
	threadID string
}

func chatInput(props chatInputProps) Node {
	return Div(
		Class("chat-input"),
		If(props.threadID != "",
			Data("signals:thread-id", fmt.Sprintf("'%s'", props.threadID)),
		),
		Data("signals:is-streaming", "false"),
		Data("signals:message", "''"),
		Data("computed:can-submit", "$message.trim() !== ''"),
		Div(
			Class("chat-input-box"),
			Data("class:is-sending", "$sendingMessage"),
			Textarea(
				AutoFocus(),
				Name("message"),
				Placeholder("Type your message here..."),
				Data("indicator:sending-message", ""),
				Data("bind:message", ""),
				Data("on:keydown", `
					if (evt.key !== 'Enter' || evt.shiftKey) return;
					if (!$canSubmit) return;
					evt.preventDefault(); // prevent writing an enter character.
					@post('/chat/send');
				`),
			),
			Div(Class("chat-input-actions"),
				Button(Class("chat-input-submit-btn"),
					Data("show", "!$isStreaming"),
					Disabled(),
					Type("submit"),
					Data("attr:disabled", "!$canSubmit || $isStreaming"),
					Data("on:click", `
						@post('/chat/send')
					`),
					iconArrowUp(),
				),
				Button(Class("chat-input-submit-btn hidden"),
					Data("class:hidden", "!$isStreaming"),
					Type("submit"),
					Data("on:click", `
						@post('/chat/stop')
					`),
					iconSquare(),
				),
			),
		),
	)
}

func existingThreadChatInput(threadID string) Node {
	return chatInput(chatInputProps{
		threadID: threadID,
	})
}

func newThreadChatInput() Node {
	return chatInput(chatInputProps{""})
}

func userMessage(messageID string, content string) Node {
	return Div(
		ID("msg-"+messageID),
		Class("message user-message"),
		Text(content),
	)
}

func RenderAssistantMessage(w io.Writer, messageID string, contentHTML string) {
	renderNode(w, assistantMessage(messageID, contentHTML, false))
}

func RenderAssistantMessageAsLast(w io.Writer, messageID string, contentHTML string) {
	renderNode(w, assistantMessage(messageID, contentHTML, true))
}

func RenderUserMessage(w io.Writer, messageID string, contentHTML string) {
	renderNode(w, userMessage(messageID, contentHTML))
}

func assistantMessage(messageID string, contentHTML string, asLast bool) Node {
	return Div(
		ID("msg-"+messageID),
		components.Classes{
			"message assistant-message": true,
			"last":                      asLast,
		},
		Raw(contentHTML),
	)
}

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
