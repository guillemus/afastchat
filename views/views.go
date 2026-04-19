package views

import (
	"fmt"
	"io"
	"log/slog"
	"net/url"

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
	renderNode(w, x.page(
		sidebar(data.Threads, "", doScrollSidebar),
		threadContentsEmpty(),
	))
}

func (x *Renderer) RenderThread(w io.Writer, data ThreadData, doScrollSidebar bool) {
	renderNode(w, x.page(
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

func sidebar(threads []Thread, activeThreadID string, doScrollSidebar bool) Node {
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
				Map(threads, func(thread Thread) Node {
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

func threadListItem(thread Thread, isActive bool) Node {
	className := "thread-item"
	if isActive {
		className += " active"
	}

	return A(
		ID("thread-"+thread.ID),
		Class(className),
		Href(fmt.Sprintf("/threads/%s", thread.ID)),

		Span(Class("thread-item-text"), Text(thread.Title)),
		If(thread.IsStreaming,
			Span(Class("thread-item-spinner"), iconSpinner()),
		),
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
	return threadContents(nil, "", false, false)
}

func threadContents(messages []Message, threadID string, doScrollSidebar bool, isStreaming bool) Node {
	var contents Node
	if len(messages) == 0 {
		input := newThreadChatInput()
		if threadID != "" {
			input = existingThreadChatInput(threadID, isStreaming)
		}

		contents = Group{
			RenderThreadMessagesNode(messages, doScrollSidebar),
			input,
		}
	} else {
		contents = Group{
			Div(
				Class("thread-shell"),
				Data("signals:show-jump-to-bottom", "false"),
				Data("on:resize__window__debounce.50ms", `
					if (!$messages) return;
					const overflow = $messages.scrollHeight - $messages.clientHeight;
					const distanceFromBottom = overflow - $messages.scrollTop;
					$showJumpToBottom = overflow > 200 && distanceFromBottom > 48;
				`),
				RenderThreadMessagesNode(messages, doScrollSidebar),
				Button(
					Type("button"),
					Style("display: none"),
					Class("jump-to-bottom-btn"),
					Data("show", "$showJumpToBottom"),
					Data("on:click", `
						$messages.scrollTo({ top: $messages.scrollHeight, behavior: 'smooth' });
						$showJumpToBottom = false;
					`),
					iconArrowDown(),
				),
			),
			existingThreadChatInput(threadID, isStreaming),
		}
	}

	return Main(Class("main"), contents)
}

func RenderThreadMessages(w io.Writer, messages []Message, doScrollSidebar bool) {
	renderNode(w, RenderThreadMessagesNode(messages, doScrollSidebar))
}

func lastAssistantMessageIndex(messages []Message) int {
	if len(messages) == 0 {
		return -1
	}

	last := len(messages) - 1
	if isAssistantMessageType(messages[last].Type) {
		return last
	}

	return -1
}

func renderMessages(messages []Message) []Node {
	lastAssistantIdx := lastAssistantMessageIndex(messages)
	out := make([]Node, 0, len(messages))
	for i, message := range messages {
		out = append(out, renderMessage(message, i == lastAssistantIdx))
	}
	return out
}

func RenderThreadMessagesNode(messages []Message, doScrollSidebar bool) Node {
	if len(messages) == 0 {
		return Div(
			ID("messages"),
			Class("main-body empty-thread"),
			Div(
				Class("empty-state"),
				H1(Class("empty-state-title"), Text("How can I help?")),
				P(Class("empty-state-subtitle"),
					Text("Start a new conversation with afastchat."),
				),
			),
		)
	}

	return Div(
		ID("messages"),
		Data("ref:messages", ""),
		components.Classes{
			"main-body": true,
			"invisible": doScrollSidebar,
		},
		Data("on:scroll__debounce.10ms", `
			const overflow = el.scrollHeight - el.clientHeight;
			const distanceFromBottom = overflow - el.scrollTop;
			$showJumpToBottom = overflow > 200 && distanceFromBottom > 48;
		`),
		If(doScrollSidebar, Data("init", `
			el.scrollTop = el.scrollHeight;
			el.classList.remove('invisible');
			const overflow = el.scrollHeight - el.clientHeight;
			const distanceFromBottom = overflow - el.scrollTop;
			$showJumpToBottom = overflow > 200 && distanceFromBottom > 48;
		`)),
		Group(renderMessages(messages)),
	)
}

type chatInputProps struct {
	threadID    string
	isStreaming bool
}

func chatInput(props chatInputProps) Node {
	return Div(
		Class("chat-input"),
		If(props.threadID != "",
			Data("signals:thread-id", fmt.Sprintf("'%s'", props.threadID)),
		),
		Data("signals:is-streaming", fmt.Sprintf("%t", props.isStreaming)),
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

func existingThreadChatInput(threadID string, isStreaming bool) Node {
	return chatInput(chatInputProps{
		threadID:    threadID,
		isStreaming: isStreaming,
	})
}

func newThreadChatInput() Node {
	return chatInput(chatInputProps{})
}

func isAssistantMessageType(messageType string) bool {
	switch messageType {
	case "assistant_message", "assistant_reasoning", "assistant_tool_call":
		return true
	default:
		return false
	}
}

func renderMessage(message Message, isLastAssistant bool) Node {
	switch message.Type {
	case "user_message":
		return userMessage(message.ID, message.MdContent)
	case "assistant_message":
		return assistantMessage(message.ID, message.MdContent, isLastAssistant)
	case "assistant_reasoning":
		return assistantReasoningMessage(message.ID, message.MdContent, isLastAssistant)
	case "assistant_tool_call":
		return assistantToolCallMessage(message.ID, message.ToolName, message.MdContent, isLastAssistant)
	default:
		return nil
	}
}

func userMessage(messageID string, content string) Node {
	return Div(
		ID("msg-"+messageID),
		Class("message user-message"),
		Text(content),
	)
}

func assistantMessage(messageID string, contentHTML string, isLastAssistant bool) Node {
	return Div(
		ID("msg-"+messageID),
		components.Classes{
			"message":                true,
			"assistant-message":      true,
			"assistant-message-last": isLastAssistant,
		},
		Raw(contentHTML),
	)
}

func assistantReasoningMessage(messageID string, contentHTML string, isLastAssistant bool) Node {
	return Div(
		ID("msg-"+messageID),
		components.Classes{
			"message":                true,
			"assistant-message":      true,
			"assistant-message-last": isLastAssistant,
		},
		Div(Text("Reasoning")),
		Raw(contentHTML),
	)
}

func assistantToolCallMessage(messageID string, toolName string, content string, isLastAssistant bool) Node {
	return Div(
		ID("msg-"+messageID),
		components.Classes{
			"message":                true,
			"assistant-message":      true,
			"assistant-message-last": isLastAssistant,
		},
		Div(Text("Tool: "+toolName)),
		Pre(Text(content)),
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
