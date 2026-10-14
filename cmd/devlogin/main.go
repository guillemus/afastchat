package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"afastchat.com/config"
	"afastchat.com/db"
	"afastchat.com/openrouter"

	"github.com/alexedwards/scs/v2"
)

const sessionLifetime = 30 * 24 * time.Hour

var testEmail = regexp.MustCompile(`^[a-z0-9._-]+@afastchat[.]test$`)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: devlogin <email>")
	}
	email := os.Args[1]
	if !testEmail.MatchString(email) {
		return fmt.Errorf("invalid test email %q (must be *@afastchat.test)", email)
	}

	cfg := config.ParseConfig()
	if !config.Dev {
		return fmt.Errorf("DEV is not true; refusing to mint a session")
	}

	ctx := context.Background()
	if err := db.Migrate(ctx, cfg.DBPath); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	pool, err := db.Connect(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	user, err := pool.LoginWithIdentity(ctx, db.LoginWithIdentityParams{
		Provider:    db.AuthProviderGoogle,
		Subject:     "dev:" + email,
		Email:       email,
		DisplayName: strings.Split(email, "@")[0],
	})
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}

	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		keys, err := openrouter.NewKeyCipher(cfg.OpenRouterTokenKey)
		if err != nil {
			return fmt.Errorf("openrouter key cipher: %w", err)
		}
		if err := pool.SaveOpenRouterConnection(ctx, user.ID, keys.Seal(apiKey)); err != nil {
			return err
		}
	} else {
		fmt.Fprintln(os.Stderr, "OPENROUTER_API_KEY is not set; the user will land on /setup")
	}

	expiry := time.Now().Add(sessionLifetime)
	data, err := scs.GobCodec{}.Encode(expiry, map[string]any{"user_id": user.ID})
	if err != nil {
		return fmt.Errorf("encode session: %w", err)
	}
	token := rand.Text()
	if err := db.NewSCSStoreWithCleanupInterval(pool, 0).CommitCtx(ctx, token, data, expiry); err != nil {
		return fmt.Errorf("save session: %w", err)
	}

	fmt.Fprintf(os.Stderr, "session for %s (user_id %s), valid 30 days\n", email, user.ID)
	fmt.Println(token)
	return nil
}
