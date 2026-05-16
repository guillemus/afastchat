package server

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"afastchat.com"
	"afastchat.com/db"
	"afastchat.com/env"
	"afastchat.com/polar"
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
	polar                 *polar.Polar
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
		polar:                 polar.New(env.PolarAPIKey),
		sessions:              sessionManager,
		crossOriginProtection: newCrossOriginProtection(),
		authIPLimiter:         newAuthRateLimiter(),
		authEmailLimiter:      newAuthRateLimiter(),
		port:                  env.Port,
	}

	s.router.Use(newSecureMiddleware(env.Dev).Handler)
	s.router.Use(s.crossOriginProtection.Handler)
	s.router.Use(s.sessions.LoadAndSave)

	// TODO: how does cloudflare interact with this? Should we even handle static asset caching ourselves, or does cf do it automatically?
	s.router.
		With(middleware.Maybe(
			middleware.NoCache,
			func(*http.Request) bool { return env.Dev },
		)).
		Get("/public/*", afastchat.GetPublicHandler(env).ServeHTTP)

	s.router.Post("/webhooks/polar", s.handlePolarWebhook)

	s.router.Get("/login", s.handleLoginPage)
	s.router.Post("/login", s.handleLogin)
	s.router.Get("/signup", s.handleSignupPage)
	s.router.Post("/signup", s.handleSignup)

	s.router.Group(func(r chi.Router) {
		r.Use(s.requireAuth)

		r.Get("/settings", s.handleSettings)
		r.Post("/settings/checkout/pro", s.handleProCheckout)
		r.Get("/settings/thank-you", s.handleCheckoutThankYou)
		r.Post("/logout", s.handleLogout)

		r.Group(func(r chi.Router) {
			r.Use(s.requireLifetimeLicense)

			r.Post("/settings/checkout/balance", s.handleBalanceCheckout)

			r.Get("/", s.handleIndex)
			r.Get("/threads/{id}", s.handleThread)
			r.Get("/threads/{id}/events", s.handleThreadEvents)
			r.Post("/threads/{id}/archive", s.handleArchiveThread)
			r.Get("/threads/search", s.handleSearchThreads)
			r.Post("/chat/send", s.handleSendMessage)
			r.Post("/chat/stop", s.handleStopStreaming)
		})
	})

	go s.resumeStreamingThreads()

	addr := ":" + s.port
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, s.router); err != nil {
		log.Fatal(err)
	}
}
