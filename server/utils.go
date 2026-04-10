package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func recoverAndLog() {
	if p := recover(); p != nil {
		slog.Error("panic recovered in goroutine", "panic", p)
	}
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

type agentCanceller struct {
	cancelMap *sync.Map
}

func newAgentCanceller() *agentCanceller {
	return &agentCanceller{
		cancelMap: &sync.Map{},
	}
}

func (x *agentCanceller) get(threadID string) (context.CancelFunc, bool) {
	val, ok := x.cancelMap.Load(threadID)
	return val.(context.CancelFunc), ok
}

func (x *agentCanceller) set(threadID string, cancel context.CancelFunc) {
	x.cancelMap.Store(threadID, cancel)
}

func (x *agentCanceller) clear(threadID string) {
	x.cancelMap.Delete(threadID)
}
