package views

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

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
			ThreadMessages(messages, doScrollSidebar),
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
				ThreadMessages(messages, doScrollSidebar),
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

func ThreadMessages(messages []Message, doScrollSidebar bool) Node {
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
		Classes{
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
		Map(messages, func(m Message) Node {
			return renderMessage(m, false)
		}),
	)
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
		Classes{
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
		Classes{
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
		Classes{
			"message":                true,
			"assistant-message":      true,
			"assistant-message-last": isLastAssistant,
		},
		Div(Text("Tool: "+toolName)),
		Pre(Text(content)),
	)
}
