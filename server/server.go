package server

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"time"

	"afastchat.com"
	"afastchat.com/config"
	"afastchat.com/db"
	"afastchat.com/openrouter"
	"afastchat.com/views"

	"github.com/alexedwards/scs/v2"
	"github.com/asaskevich/EventBus"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
)

type Server struct {
	cfg                   config.Config
	bus                   EventBus.Bus
	router                chi.Router
	db                    *db.Pool
	sessions              *scs.SessionManager
	crossOriginProtection *http.CrossOriginProtection
	googleOAuth           oauth2.Config
	googleVerifier        *oidc.IDTokenVerifier
	openRouterKeys        *openrouter.KeyCipher
	port                  string
}

func StartServer(cfg config.Config) {
	if err := db.Migrate(context.Background(), cfg.DBPath); err != nil {
		log.Fatalf("db: %v", err)
	}

	conn, err := db.Connect(cfg.DBPath)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	openRouterKeys, err := openrouter.NewKeyCipher(cfg.OpenRouterTokenKey)
	if err != nil {
		log.Fatalf("openrouter: %v", err)
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
		sessions:              sessionManager,
		crossOriginProtection: newCrossOriginProtection(),
		googleOAuth: oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:   "https://accounts.google.com/o/oauth2/auth",
				TokenURL:  "https://oauth2.googleapis.com/token",
				AuthStyle: oauth2.AuthStyleInParams,
			},
			RedirectURL: config.AppURL() + "/auth/google/callback",
			Scopes:      []string{oidc.ScopeOpenID, "email", "profile"},
		},
		openRouterKeys: openRouterKeys,
		googleVerifier: oidc.NewVerifier(
			"https://accounts.google.com",
			oidc.NewRemoteKeySet(context.Background(), "https://www.googleapis.com/oauth2/v3/certs"),
			&oidc.Config{ClientID: cfg.GoogleClientID},
		),
		port: cfg.Port,
	}

	s.router.Use(newSecureMiddleware(config.Dev).Handler)
	s.router.Use(s.crossOriginProtection.Handler)
	s.router.Use(s.sessions.LoadAndSave)
	s.router.Use(withTimeZone)

	s.router.Handle("/public/*", afastchat.PublicHandler())
	s.router.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/public/favicon.ico", http.StatusMovedPermanently)
	})

	s.router.Get("/", s.handleRoot)
	s.router.Get("/login", s.handleLoginPage)
	s.router.Get("/privacy", func(w http.ResponseWriter, r *http.Request) {
		s.render(w, r, views.PrivacyPage())
	})
	s.router.Get("/terms", func(w http.ResponseWriter, r *http.Request) {
		s.render(w, r, views.TermsPage())
	})
	s.router.Post("/auth/google", s.handleGoogleLogin)
	s.router.Get("/auth/google/callback", s.handleGoogleCallback)

	s.router.Group(func(r chi.Router) {
		r.Use(s.requireAuth)

		r.Get("/setup", s.handleSetup)
		r.Post("/auth/openrouter", s.handleOpenRouterConnect)
		r.Get("/auth/openrouter/callback", s.handleOpenRouterCallback)
		r.Get("/settings", s.handleSettings)
		r.Post("/settings/openrouter/disconnect", s.handleOpenRouterDisconnect)
		r.Post("/settings/delete-account", s.handleDeleteAccount)
		r.Post("/logout", s.handleLogout)
		r.Get("/threads/{id}/events", s.handleThreadEvents)
		r.Post("/threads/{id}/archive", s.handleArchiveThread)
		r.Post("/threads/{id}/rename", s.handleRenameThread)
		r.Post("/threads/{id}/delete", s.handleDeleteThread)
		r.Get("/threads/search", s.handleSearchThreads)
		r.Post("/chat/stop", s.handleStopStreaming)
		r.Post("/user/selected-model", s.handleSelectModel)
		r.Post("/user/favorite-models", s.handleFavoriteModel)

		r.Group(func(r chi.Router) {
			r.Use(s.requireOpenRouter)

			r.Get("/threads/{id}", s.handleThread)
			r.Post("/chat/send", s.handleSendMessage)
		})
	})

	go s.resumeStreamingThreads()

	addr := ":" + s.port
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, s.router); err != nil {
		log.Fatal(err)
	}
}
