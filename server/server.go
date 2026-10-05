package server

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"afastchat.com"
	"afastchat.com/config"
	"afastchat.com/db"
	"afastchat.com/polar"

	"github.com/alexedwards/scs/v2"
	"github.com/asaskevich/EventBus"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
)

type Server struct {
	cfg                   config.Config
	bus                   EventBus.Bus
	router                chi.Router
	db                    *db.Pool
	polar                 *polar.Polar
	sessions              *scs.SessionManager
	crossOriginProtection *http.CrossOriginProtection
	authIPLimiter         *httprate.RateLimiter
	authEmailLimiter      *httprate.RateLimiter
	port                  string
}

func StartServer(cfg config.Config) {
	conn, err := db.Connect(cfg.DBPath)
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
	sessionManager.Cookie.Secure = !config.Dev

	s := &Server{
		cfg:                   cfg,
		bus:                   EventBus.New(),
		router:                chi.NewRouter(),
		db:                    conn,
		polar:                 polar.New(cfg.PolarAPIKey),
		sessions:              sessionManager,
		crossOriginProtection: newCrossOriginProtection(),
		authIPLimiter:         newAuthRateLimiter(),
		authEmailLimiter:      newAuthRateLimiter(),
		port:                  cfg.Port,
	}

	s.router.Use(newSecureMiddleware(config.Dev).Handler)
	s.router.Use(s.crossOriginProtection.Handler)
	s.router.Use(s.sessions.LoadAndSave)

	s.router.Handle("/public/*", afastchat.PublicHandler())

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
