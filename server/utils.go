package server

import (
	"log/slog"
	"net/http"
	"time"

	"afastchat.com/db"
	"afastchat.com/views"
	"afastchat.com/views/ui"
	"github.com/a-h/templ"
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
		IsPinned:    thread.PinnedAt != nil,
		Title:       thread.Title,
		CreatedAt:   thread.CreatedAt,
	}
}

func mapViewThreads(threads []db.Thread) []views.Thread {
	out := make([]views.Thread, 0, len(threads))
	for _, thread := range threads {
		out = append(out, viewThreadFromDbThread(thread))
	}
	return out
}

func (x *Server) render(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := component.Render(r.Context(), w); err != nil {
		slog.Error("render component", "err", err)
	}
}

func (x *Server) patchToast(w http.ResponseWriter, r *http.Request, toast ui.Toast) {
	sse := datastar.NewSSE(w, r)
	patchToastOnSSE(sse, toast)
}

func patchToastOnSSE(sse *datastar.ServerSentEventGenerator, toast ui.Toast) {
	toast.ID = time.Now().UnixNano()
	toast.Visible = true

	views.PatchSignals(sse, map[string]any{"_toast": toast})
}
