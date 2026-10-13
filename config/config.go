package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

var (
	AssetVersion = strconv.FormatInt(time.Now().UnixNano(), 10)
	Dev          bool
)

type Config struct {
	LogLevel           string
	DBPath             string
	Port               string
	OpenRouterAPIKey   string
	GoogleClientID     string
	GoogleClientSecret string
}

func ParseConfig() Config {
	_ = godotenv.Load()

	Dev = os.Getenv("DEV") == "true"

	return Config{
		LogLevel:           GetEnvOr("LOG_LEVEL", "info"),
		DBPath:             GetEnvOr("DB_PATH", "./tmp/app.db"),
		Port:               GetEnvOr("PORT", "3000"),
		OpenRouterAPIKey:   os.Getenv("OPENROUTER_API_KEY"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	}
}

func AppURL() string {
	if Dev {
		return "https://localhost"
	}
	return "https://afastchat.com"
}

func GetEnvOr(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
