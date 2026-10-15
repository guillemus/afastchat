package server

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"afastchat.com/config"
	"afastchat.com/db"
	"afastchat.com/views"
	"afastchat.com/views/ui"

	"github.com/starfederation/datastar-go/datastar"
	"golang.org/x/oauth2"
)

const (
	oauthStateCookieName    = "afastchat_oauth_state"
	oauthNonceCookieName    = "afastchat_oauth_nonce"
	oauthVerifierCookieName = "afastchat_oauth_verifier"

	oauthTransactionDuration = 10 * time.Minute
)

func (x *Server) currentUser(r *http.Request) (user db.User, found bool, err error) {
	ctx := r.Context()

	userID := x.sessions.GetString(ctx, "user_id")
	if userID == "" {
		return user, false, nil
	}

	user, err = x.db.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			if err := x.sessions.Destroy(ctx); err != nil {
				return user, false, err
			}
			return user, false, nil
		}
		return user, false, err
	}

	return user, true, nil
}

type userContextKey struct{}

func (x *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, found, err := x.currentUser(r)
		if err != nil {
			slog.Error("load current user", "err", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if !found {
			if r.Method == http.MethodGet && isHTMLNavigation(r) {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			} else if isDatastar(r) {
				x.patchToast(w, r, ui.Toast{
					Kind:    ui.ToastWarning,
					Message: "Sign in, then try that again.",
					Link:    ui.ToastLink{Label: "Sign in", Href: "/login"},
				})
			} else {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
			}
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func mustGetUser(r *http.Request) db.User {
	user, ok := r.Context().Value(userContextKey{}).(db.User)
	if !ok {
		panic("user missing from request context")
	}
	return user
}

func (x *Server) landingForGuests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if x.sessions.GetString(r.Context(), "user_id") == "" {
			x.render(w, r, views.LandingPage())
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (x *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	_, found, err := x.currentUser(r)
	if err != nil {
		slog.Error("login page: current user", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if found {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	x.render(w, r, views.LoginPage(r.URL.Query().Get("failed") == "1"))
}

func (x *Server) createSession(r *http.Request, userID string) error {
	ctx := r.Context()
	if err := x.sessions.RenewToken(ctx); err != nil {
		return err
	}
	x.sessions.Put(ctx, "user_id", userID)
	return nil
}

func (x *Server) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := randomToken(32)
	if err != nil {
		slog.Error("generate OAuth state", "err", err)
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastError, Message: "Login could not be started. Try again."})
		return
	}

	nonce, err := randomToken(32)
	if err != nil {
		slog.Error("generate OIDC nonce", "err", err)
		x.patchToast(w, r, ui.Toast{Kind: ui.ToastError, Message: "Login could not be started. Try again."})
		return
	}
	verifier := oauth2.GenerateVerifier()

	setOAuthTransactionCookie(w, oauthStateCookieName, state)
	setOAuthTransactionCookie(w, oauthNonceCookieName, nonce)
	setOAuthTransactionCookie(w, oauthVerifierCookieName, verifier)

	authURL := x.googleOAuth.AuthCodeURL(
		state,
		oauth2.S256ChallengeOption(verifier),
		oauth2.SetAuthURLParam("nonce", nonce),
	)
	sse := datastar.NewSSE(w, r)
	if err := sse.Redirect(authURL); err != nil {
		slog.Error("redirect to Google", "err", err)
	}
}

type googleClaims struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Nonce         string `json:"nonce"`
}

func (x *Server) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	transaction, err := readOAuthTransaction(r)
	clearCookies(w, oauthStateCookieName, oauthNonceCookieName, oauthVerifierCookieName)
	if err != nil {
		googleCallbackFailed(w, r, err)
		return
	}
	if providerError := r.URL.Query().Get("error"); providerError != "" {
		googleCallbackFailed(w, r, fmt.Errorf("provider returned %s", providerError))
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		googleCallbackFailed(w, r, fmt.Errorf("authorization code is missing"))
		return
	}

	token, err := x.googleOAuth.Exchange(ctx, code, oauth2.VerifierOption(transaction.verifier))
	if err != nil {
		googleCallbackFailed(w, r, fmt.Errorf("exchange authorization code: %w", err))
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		googleCallbackFailed(w, r, fmt.Errorf("ID token is missing"))
		return
	}
	idToken, err := x.googleVerifier.Verify(ctx, rawIDToken)
	if err != nil {
		googleCallbackFailed(w, r, fmt.Errorf("verify ID token: %w", err))
		return
	}

	var claims googleClaims
	if err := idToken.Claims(&claims); err != nil {
		googleCallbackFailed(w, r, fmt.Errorf("decode ID token claims: %w", err))
		return
	}
	if claims.Subject == "" || claims.Email == "" || !claims.EmailVerified ||
		claims.Nonce == "" || !equalSecret(claims.Nonce, transaction.nonce) {
		googleCallbackFailed(w, r, fmt.Errorf("invalid ID token claims"))
		return
	}

	displayName := strings.TrimSpace(claims.Name)
	if displayName == "" {
		displayName = strings.Split(claims.Email, "@")[0]
	}

	user, err := x.db.LoginWithIdentity(ctx, db.LoginWithIdentityParams{
		Provider:    db.AuthProviderGoogle,
		Subject:     claims.Subject,
		Email:       claims.Email,
		DisplayName: displayName,
	})
	if err != nil {
		googleCallbackFailed(w, r, fmt.Errorf("complete google login: %w", err))
		return
	}

	if err := x.createSession(r, user.ID); err != nil {
		googleCallbackFailed(w, r, fmt.Errorf("create session: %w", err))
		return
	}

	slog.Info("user logged in", "user_id", user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (x *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := x.sessions.Destroy(ctx); err != nil {
		slog.Error("logout: destroy session", "err", err)
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func googleCallbackFailed(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("Google authentication failed", "err", err)
	http.Redirect(w, r, "/login?failed=1", http.StatusSeeOther)
}

func setOAuthTransactionCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name: name, Value: value, Path: "/",
		MaxAge: int(oauthTransactionDuration.Seconds()), HttpOnly: true,
		Secure: !config.Dev, SameSite: http.SameSiteLaxMode,
	})
}

func clearCookies(w http.ResponseWriter, names ...string) {
	for _, name := range names {
		http.SetCookie(w, &http.Cookie{
			Name: name, Value: "", Path: "/", MaxAge: -1, HttpOnly: true,
			Secure: !config.Dev, SameSite: http.SameSiteLaxMode,
		})
	}
}

type oauthTransaction struct {
	nonce    string
	verifier string
}

func readOAuthTransaction(r *http.Request) (oauthTransaction, error) {
	state := r.URL.Query().Get("state")
	if state == "" {
		return oauthTransaction{}, errors.New("OAuth state is missing")
	}

	stateCookie, err := r.Cookie(oauthStateCookieName)
	if err != nil {
		return oauthTransaction{}, errors.New("OAuth state cookie is missing")
	}

	nonceCookie, err := r.Cookie(oauthNonceCookieName)
	if err != nil {
		return oauthTransaction{}, errors.New("OIDC nonce cookie is missing")
	}

	verifierCookie, err := r.Cookie(oauthVerifierCookieName)
	if err != nil {
		return oauthTransaction{}, errors.New("PKCE verifier cookie is missing")
	}

	if !equalSecret(state, stateCookie.Value) {
		return oauthTransaction{}, errors.New("OAuth state does not match")
	}

	if nonceCookie.Value == "" {
		return oauthTransaction{}, errors.New("OIDC nonce is empty")
	}

	if verifierCookie.Value == "" {
		return oauthTransaction{}, errors.New("PKCE verifier is empty")
	}

	return oauthTransaction{
		nonce:    nonceCookie.Value,
		verifier: verifierCookie.Value,
	}, nil
}

func equalSecret(left, right string) bool {
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}

func randomToken(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}
