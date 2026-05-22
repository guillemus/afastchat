package server

import (
	"log/slog"
	"net/http"

	"afastchat.com/db"
	"afastchat.com/views"
	"github.com/starfederation/datastar-go/datastar"
)

func GoRecover(fn func()) {
	go func() {
		defer func() {
			if p := recover(); p != nil {
				slog.Error("panic recovered in goroutine", "panic", p)
			}
		}()
		fn()
	}()
}

func handleBadReq(w http.ResponseWriter, msg string, err error) {
	slog.Error(msg, "err", err)
	http.Error(w, "bad request", http.StatusBadRequest)
}

func viewThreadFromDbThread(thread db.Thread) views.Thread {
	return views.Thread{
		ID:          thread.ID,
		IsStreaming: thread.IsStreaming,
		Title:       thread.Title,
	}
}

func mapViewThreads(threads []db.Thread) []views.Thread {
	out := make([]views.Thread, 0, len(threads))
	for _, thread := range threads {
		out = append(out, viewThreadFromDbThread(thread))
	}
	return out
}

func mapViewMessages(messages []db.Message, largeHeight bool) []views.MessageGroup {
	out := make([]views.MessageGroup, 0, len(messages))
	var assistant *views.AssistantMessagePartsProps

	for _, it := range messages {
		switch it.Type {
		case db.SystemMessage, db.ToolOutput:
			continue
		case db.UserMessage:
			if assistant != nil {
				out = append(out, *assistant)
				assistant = nil
			}
			out = append(out, views.UserMessageProps{
				ID:        it.ID,
				MdContent: renderMarkdown(it.Content),
			})
		case db.AssistantMessage:
			if assistant == nil {
				assistant = &views.AssistantMessagePartsProps{
					ID: it.ID,
				}
			}
			assistant.Parts = append(assistant.Parts, views.AssistantTextPartProps{
				ID:        it.ID,
				MdContent: renderMarkdown(it.Content),
			})
		case db.AssistantReasoning:
			if assistant == nil {
				assistant = &views.AssistantMessagePartsProps{
					ID: it.ID,
				}
			}
			assistant.Parts = append(assistant.Parts, views.ReasoningPartProps{
				ID:        it.ID,
				Text:      it.Content,
				MdContent: renderMarkdown(it.Content),
			})
		case db.AssistantToolCall:
			if assistant == nil {
				assistant = &views.AssistantMessagePartsProps{
					ID: it.ID,
				}
			}
			assistant.Parts = append(assistant.Parts, views.ToolCallPartProps{
				ID:   it.ID,
				Name: it.ToolName,
				Args: it.Content,
			})
		}
	}

	if assistant != nil {
		out = append(out, *assistant)
		assistant = nil
	}

	if largeHeight {
		for i := len(out) - 1; i >= 0; i-- {
			if msg, ok := out[i].(views.AssistantMessagePartsProps); ok {
				msg.LargeHeight = true
				out[i] = msg
				break
			}
		}
	}

	return out
}

func toastShowError(w http.ResponseWriter, r *http.Request, title, message string) {
	sse := datastar.NewSSE(w, r)
	views.PatchErrorToast(sse, title, message)
}

func toastShow(w http.ResponseWriter, r *http.Request, toast views.ToastData) {
	sse := datastar.NewSSE(w, r)
	views.PatchToast(sse, toast)
}
