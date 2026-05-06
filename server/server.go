package server

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"afastchat.com"
	"afastchat.com/db"
	"afastchat.com/env"
	"afastchat.com/views"

	"github.com/alexedwards/scs/v2"
	"github.com/asaskevich/EventBus"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

type Server struct {
	env                   env.Env
	bus                   EventBus.Bus
	router                chi.Router
	renderer              *views.Renderer
	db                    *db.Pool
	sessions              *scs.SessionManager
	crossOriginProtection *http.CrossOriginProtection
	authIPLimiter         *httprate.RateLimiter
	authEmailLimiter      *httprate.RateLimiter
	port                  string
}

func StartServer(env env.Env) {
	conn, err := db.Connect(env.DBPath)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	sessionManager := scs.New()
	sessionManager.Store = db.NewSCSStore(conn)
	sessionManager.Lifetime = 30 * 24 * time.Hour
	sessionManager.IdleTimeout = 7 * 24 * time.Hour
	sessionManager.Cookie.Name = "afastchat_session"
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Secure = !env.Dev

	s := &Server{
		env:                   env,
		bus:                   EventBus.New(),
		router:                chi.NewRouter(),
		renderer:              views.NewRenderer(env),
		db:                    conn,
		sessions:              sessionManager,
		crossOriginProtection: newCrossOriginProtection(),
		authIPLimiter:         newAuthRateLimiter(),
		authEmailLimiter:      newAuthRateLimiter(),
		port:                  env.Port,
	}

	s.router.Use(newSecureMiddleware(env.Dev).Handler)
	s.router.Use(s.crossOriginProtection.Handler)
	s.router.Use(s.sessions.LoadAndSave)

	s.router.
		With(middleware.Maybe(
			middleware.NoCache,
			func(*http.Request) bool { return env.Dev },
		)).
		Get("/public/*", afastchat.GetPublicHandler(env).ServeHTTP)

	s.router.Get("/", s.withUser(s.handleIndex))
	s.router.Get("/login", s.handleLoginPage)
	s.router.Post("/login", s.handleLogin)
	s.router.Get("/signup", s.handleSignupPage)
	s.router.Post("/signup", s.handleSignup)
	s.router.Post("/logout", s.handleLogout)
	s.router.Get("/threads/{id}", s.handlePathVal(s.handleThread, "id"))
	s.router.Get("/threads/{id}/events", s.handlePathVal(s.handleThreadEvents, "id"))
	s.router.Post("/threads/{id}/archive", s.handlePathVal(s.handleArchiveThread, "id"))
	s.router.Get("/threads/search", handleSignals(s, s.handleSearchThreads))
	s.router.Post("/chat/send", handleSignals(s, s.handleSendMessage))
	s.router.Post("/chat/stop", handleSignals(s, s.handleStopStreaming))

	addr := ":" + s.port
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, s.router); err != nil {
		log.Fatal(err)
	}
}
