package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"afastchat.com/config"
	"afastchat.com/db"
	"afastchat.com/openrouter"
	"afastchat.com/views"
	"afastchat.com/views/ui"

	"github.com/starfederation/datastar-go/datastar"
	"golang.org/x/oauth2"
)

const (
	openRouterStateCookieName    = "afastchat_openrouter_state"
	openRouterVerifierCookieName = "afastchat_openrouter_verifier"
	openRouterReturnCookieName   = "afastchat_openrouter_return"
)

type openRouterKeyContextKey struct{}

func (x *Server) openRouterKey(ctx context.Context, userID string) (key string, found bool, err error) {
	sealed, err := x.db.GetOpenRouterSealedKey(ctx, userID)
	if errors.Is(err, db.ErrNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}

	key, err = x.openRouterKeys.Open(sealed)
	if errors.Is(err, openrouter.ErrUnknownSealingKey) {
		slog.Warn("openrouter key sealed with a retired key", "user_id", userID)
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}

	return key, true, nil
}

func (x *Server) requireOpenRouter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := mustGetUser(r)
		key, found, err := x.openRouterKey(r.Context(), user.ID)
		if err != nil {
			slog.Error("load openrouter key", "err", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if !found {
			if isDatastar(r) {
				x.patchToast(w, r, ui.Toast{
					Kind:    ui.ToastWarning,
					Message: "Connect your OpenRouter account to start chatting.",
					Link:    ui.ToastLink{Label: "Connect", Href: "/setup"},
				})
			} else {
				http.Redirect(w, r, "/setup", http.StatusSeeOther)
			}
			return
		}

		ctx := context.WithValue(r.Context(), openRouterKeyContextKey{}, key)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func mustGetOpenRouterKey(r *http.Request) string {
	key, ok := r.Context().Value(openRouterKeyContextKey{}).(string)
	if !ok {
		panic("openrouter key missing from request context")
	}
	return key
}

func (x *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	user := mustGetUser(r)
	_, found, err := x.openRouterKey(r.Context(), user.ID)
	if err != nil {
		slog.Error("setup: load openrouter key", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if found {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	x.render(w, r, views.SetupPage(r.URL.Query().Get("failed") == "1"))
}

func (x *Server) handleOpenRouterConnect(w http.ResponseWriter, r *http.Request) {
	state, err := randomToken(32)
	if err != nil {
		slog.Error("generate openrouter state", "err", err)
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastError, Message: "OpenRouter could not be connected. Try again."})
		return
	}
	verifier := oauth2.GenerateVerifier()

	returnTo := "/"
	if r.URL.Query().Get("from") == "settings" {
		returnTo = "/settings"
	}

	setOAuthTransactionCookie(w, openRouterStateCookieName, state)
	setOAuthTransactionCookie(w, openRouterVerifierCookieName, verifier)
	setOAuthTransactionCookie(w, openRouterReturnCookieName, returnTo)

	authURL := openrouter.AuthCodeURL(config.AppURL()+"/auth/openrouter/callback", state, verifier)
	sse := datastar.NewSSE(w, r)
	if err := sse.Redirect(authURL); err != nil {
		slog.Error("redirect to openrouter", "err", err)
	}
}

func (x *Server) handleOpenRouterCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := mustGetUser(r)

	returnTo := "/"
	if cookie, err := r.Cookie(openRouterReturnCookieName); err == nil && cookie.Value == "/settings" {
		returnTo = "/settings"
	}
	failed := func(err error) {
		slog.Error("openrouter connection failed", "user_id", user.ID, "err", err)
		if returnTo == "/settings" {
			http.Redirect(w, r, "/settings?openrouter_failed=1", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/setup?failed=1", http.StatusSeeOther)
		}
	}

	stateCookie, stateErr := r.Cookie(openRouterStateCookieName)
	verifierCookie, verifierErr := r.Cookie(openRouterVerifierCookieName)
	clearCookies(w, openRouterStateCookieName, openRouterVerifierCookieName, openRouterReturnCookieName)
	if stateErr != nil || verifierErr != nil || verifierCookie.Value == "" {
		failed(errors.New("openrouter transaction cookies are missing"))
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" || !equalSecret(state, stateCookie.Value) {
		failed(errors.New("openrouter state does not match"))
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		failed(errors.New("openrouter authorization code is missing"))
		return
	}

	key, err := openrouter.ExchangeCode(ctx, code, verifierCookie.Value)
	if err != nil {
		failed(err)
		return
	}

	if err := x.db.SaveOpenRouterConnection(ctx, user.ID, x.openRouterKeys.Seal(key)); err != nil {
		failed(fmt.Errorf("save openrouter connection: %w", err))
		return
	}

	slog.Info("openrouter connected", "user_id", user.ID)
	http.Redirect(w, r, returnTo, http.StatusSeeOther)
}

func (x *Server) handleOpenRouterDisconnect(w http.ResponseWriter, r *http.Request) {
	user := mustGetUser(r)
	if err := x.db.DeleteOpenRouterConnection(r.Context(), user.ID); err != nil {
		slog.Error("disconnect openrouter", "err", err)
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastError, Message: "OpenRouter could not be disconnected. Try again."})
		return
	}

	slog.Info("openrouter disconnected", "user_id", user.ID)
	sse := datastar.NewSSE(w, r)
	if err := sse.Redirect("/settings"); err != nil {
		slog.Error("redirect to settings", "err", err)
	}
}
