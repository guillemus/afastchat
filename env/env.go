package env

import (
	"fmt"
	"os"
	"time"

	"github.com/guillemus/required"
	"github.com/joho/godotenv"
)

type Config struct {
	DBPath           string
	OpenAIAPIKey     string
	OpenRouterAPIKey string
	Port             string
	AssetVersion     string
	Dev              bool
}

var Env = parseEnv()

func parseEnv() *Config {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		panic(fmt.Sprintf("load .env file: %v", err))
	}

	cfg := &Config{
		DBPath:           getEnvOr("DB_PATH", "./tmp/app.db"),
		OpenAIAPIKey:     os.Getenv("OPENAI_API_KEY"),
		OpenRouterAPIKey: os.Getenv("OPENROUTER_API_KEY"),
		Port:             getEnvOr("PORT", "3000"),
		AssetVersion:     fmt.Sprintf("%d", time.Now().Unix()),
		Dev:              os.Getenv("DEV") == "true",
	}

	if err := required.Required(cfg); err != nil {
		panic(fmt.Sprintf("validate env vars: %v", err))
	}

	return cfg
}

func getEnvOr(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
