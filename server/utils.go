package server

import (
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

func handleSignals[T any](
	s *Server,
	fn func(w http.ResponseWriter, r *http.Request, user db.User, v T),
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var t T
		if err := datastar.ReadSignals(r, &t); err != nil {
			handleBadReq(w, "read signals failed", err)
			return
		}

		user, ok := s.requireUser(w, r)
		if !ok {
			return
		}

		slog.Debug("read signals", "signals", t)

		fn(w, r, user, t)
	}
}

func (x *Server) withUser(
	fn func(w http.ResponseWriter, r *http.Request, user db.User),
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := x.requireUser(w, r)
		if !ok {
			return
		}

		fn(w, r, user)
	}
}

func (x *Server) handlePathVal(
	fn func(w http.ResponseWriter, r *http.Request, user db.User, val string),
	pathValName string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val := chi.URLParam(r, pathValName)
		if val == "" {
			handleBadReq(w, "missing path value", nil)
			return
		}

		user, ok := x.requireUser(w, r)
		if !ok {
			return
		}

		fn(w, r, user, val)
	}
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
	var assistant *views.AssistantMessageParts

	for _, it := range messages {
		switch it.Type {
		case db.SystemMessage, db.ToolOutput:
			continue
		case db.UserMessage:
			if assistant != nil {
				out = append(out, *assistant)
				assistant = nil
			}
			out = append(out, views.UserMessage{
				ID:        it.ID,
				MdContent: renderMarkdown(it.Content),
			})
		case db.AssistantMessage:
			if assistant == nil {
				assistant = &views.AssistantMessageParts{
					ID: it.ID,
				}
			}
			assistant.Parts = append(assistant.Parts, views.AssistantTextPart{
				ID:        it.ID,
				MdContent: renderMarkdown(it.Content),
			})
		case db.AssistantReasoning:
			if assistant == nil {
				assistant = &views.AssistantMessageParts{
					ID: it.ID,
				}
			}
			assistant.Parts = append(assistant.Parts, views.ReasoningPart{
				ID:   it.ID,
				Text: renderMarkdown(it.Content),
			})
		case db.AssistantToolCall:
			if assistant == nil {
				assistant = &views.AssistantMessageParts{
					ID: it.ID,
				}
			}
			assistant.Parts = append(assistant.Parts, views.ToolCallPart{
				ID:   it.ID,
				Text: it.Content,
			})
		}
	}

	if assistant != nil {
		out = append(out, *assistant)
		assistant = nil
	}

	if largeHeight {
		for i := len(out) - 1; i >= 0; i-- {
			if msg, ok := out[i].(views.AssistantMessageParts); ok {
				msg.LargeHeight = true
				out[i] = msg
				break
			}
		}
	}

	return out
}
