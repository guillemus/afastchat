package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"

	"afastchat.com/db"
	"afastchat.com/views"

	"golang.org/x/crypto/bcrypt"
)

const (
	invalidCredentialsMessage = "Invalid email or password."
	invalidLoginInputMessage  = "Enter valid email and password."
	invalidSignupInputMessage = "Enter a valid email and a password with at least 8 characters."
	emailTakenMessage         = "An account with that email already exists."
	internalLoginErrorMessage = "Something went wrong. Try again."
	minimumPasswordLength     = 8
)

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}

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
			} else {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
			}
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (x *Server) requirePaid(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := mustGetUser(r)
		if user.PaidAt == nil {
			if r.Method == http.MethodGet && isHTMLNavigation(r) {
				http.Redirect(w, r, "/settings", http.StatusSeeOther)
			} else {
				http.Error(w, "payment required", http.StatusPaymentRequired)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

func mustGetUser(r *http.Request) db.User {
	user, ok := r.Context().Value(userContextKey{}).(db.User)
	if !ok {
		panic("user missing from request context")
	}
	return user
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

	x.renderer.RenderLoginPage(w, views.LoginPageData{})
}

func (x *Server) handleSignupPage(w http.ResponseWriter, r *http.Request) {
	_, found, err := x.currentUser(r)
	if err != nil {
		slog.Error("signup page: current user", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if found {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	x.renderer.RenderSignupPage(w, views.SignupPageData{})
}

func (x *Server) createSession(r *http.Request, userID string) error {
	ctx := r.Context()
	if err := x.sessions.RenewToken(ctx); err != nil {
		return err
	}
	x.sessions.Put(ctx, "user_id", userID)
	return nil
}

func (x *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	email := normalizeEmail(r.FormValue("email"))
	password := r.FormValue("password")

	if x.authRateLimited(w, r, email) {
		w.WriteHeader(http.StatusTooManyRequests)
		x.renderer.RenderLoginPage(w, views.LoginPageData{
			Email: email,
			Error: tooManyAttemptsMessage,
		})
		return
	}

	if !isValidEmail(email) || len(password) < minimumPasswordLength {
		x.renderer.RenderLoginPage(w, views.LoginPageData{
			Email: email,
			Error: invalidLoginInputMessage,
		})
		return
	}

	user, err := x.db.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			x.renderer.RenderLoginPage(w, views.LoginPageData{
				Email: email,
				Error: invalidCredentialsMessage,
			})
			return
		}
		slog.Error("login: get user by email", "email", email, "err", err)
		x.renderer.RenderLoginPage(w, views.LoginPageData{
			Email: email,
			Error: internalLoginErrorMessage,
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		x.renderer.RenderLoginPage(w, views.LoginPageData{
			Email: email,
			Error: invalidCredentialsMessage,
		})
		return
	}

	if err := x.createSession(r, user.ID); err != nil {
		slog.Error("login: create session", "user_id", user.ID, "err", err)
		x.renderer.RenderLoginPage(w, views.LoginPageData{
			Email: email,
			Error: internalLoginErrorMessage,
		})
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (x *Server) handleSignup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	email := normalizeEmail(r.FormValue("email"))
	password := r.FormValue("password")

	if x.authRateLimited(w, r, email) {
		w.WriteHeader(http.StatusTooManyRequests)
		x.renderer.RenderSignupPage(w, views.SignupPageData{
			Email: email,
			Error: tooManyAttemptsMessage,
		})
		return
	}

	if !isValidEmail(email) || len(password) < minimumPasswordLength {
		x.renderer.RenderSignupPage(w, views.SignupPageData{
			Email: email,
			Error: invalidSignupInputMessage,
		})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("signup: hash password", "err", err)
		x.renderer.RenderSignupPage(w, views.SignupPageData{
			Email: email,
			Error: internalLoginErrorMessage,
		})
		return
	}

	user, err := x.db.CreateUser(ctx, email, string(hash))
	if err != nil {
		if errors.Is(err, db.ErrDuplicateEmail) {
			x.renderer.RenderSignupPage(w, views.SignupPageData{
				Email: email,
				Error: emailTakenMessage,
			})
			return
		}
		slog.Error("signup: create user", "email", email, "err", err)
		x.renderer.RenderSignupPage(w, views.SignupPageData{
			Email: email,
			Error: internalLoginErrorMessage,
		})
		return
	}

	if err := x.createSession(r, user.ID); err != nil {
		slog.Error("signup: create session", "user_id", user.ID, "err", err)
		x.renderer.RenderSignupPage(w, views.SignupPageData{
			Email: email,
			Error: internalLoginErrorMessage,
		})
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (x *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := x.sessions.Destroy(ctx); err != nil {
		slog.Error("logout: destroy session", "err", err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
