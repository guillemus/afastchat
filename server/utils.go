package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"afastchat.com/db"
	"afastchat.com/views"
	"github.com/go-chi/chi/v5"
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

func handleSignals[T any](fn func(w http.ResponseWriter, r *http.Request, v T)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var t T
		if err := datastar.ReadSignals(r, &t); err != nil {
			handleBadReq(w, "read signals failed", err)
			return
		}

		slog.Debug("read signals", "signals", t)

		fn(w, r, t)
	}
}

func handlePathVal(
	fn func(w http.ResponseWriter, r *http.Request, val string),
	pathValName string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val := chi.URLParam(r, pathValName)
		if val == "" {
			handleBadReq(w, "missing path value", nil)
			return
		}

		fn(w, r, val)
	}
}

func patchSignals(sse *datastar.ServerSentEventGenerator, v any) {
	bs, err := json.Marshal(v)
	if err != nil {
		slog.Error("failed to marshal", "value", v)
	}

	sse.PatchSignals(bs)
}

func mapViewThread(thread db.Thread) views.Thread {
	return views.Thread{
		ID:          thread.ID,
		IsStreaming: thread.IsStreaming,
		Title:       thread.Title,
	}
}

func mapViewThreads(threads []db.Thread) []views.Thread {
	out := make([]views.Thread, 0, len(threads))
	for _, thread := range threads {
		out = append(out, mapViewThread(thread))
	}
	return out
}

func mapViewMessages(messages []db.Message) []views.Message {
	out := make([]views.Message, 0, len(messages))
	for _, message := range messages {
		viewMessage := views.Message{
			Type:      message.Type,
			ID:        message.ID,
			MdContent: message.Content,
			ToolName:  message.ToolName,
		}

		switch message.Type {
		case "system_message", "tool_output":
			continue
		case "assistant_message", "assistant_reasoning":
			viewMessage.MdContent = renderMarkdown(message.Content)
		}

		out = append(out, viewMessage)
	}
	return out
}
