package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

func handleUUIDPathVal(
	fn func(w http.ResponseWriter, r *http.Request, val uuid.UUID),
	pathValName string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, err := uuid.Parse(chi.URLParam(r, pathValName))
		if err != nil {
			handleBadReq(w, "invalid uuid", err)
			return
		}

		fn(w, r, val)
	}
}
