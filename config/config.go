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
	OpenAIAPIKey       string
	FireworksAPIKey    string
	ExaAPIKey          string
	PolarAPIKey        string
	PolarWebhookSecret string
}

func ParseConfig() Config {
	_ = godotenv.Load()

	Dev = os.Getenv("DEV") == "true"

	return Config{
		LogLevel:           GetEnvOr("LOG_LEVEL", "info"),
		DBPath:             GetEnvOr("DB_PATH", "./tmp/app.db"),
		Port:               GetEnvOr("PORT", "3000"),
		OpenAIAPIKey:       os.Getenv("OPENAI_API_KEY"),
		FireworksAPIKey:    os.Getenv("FIREWORKS_API_KEY"),
		ExaAPIKey:          os.Getenv("EXA_AI_API_KEY"),
		PolarAPIKey:        os.Getenv("POLAR_ACCESS_TOKEN"),
		PolarWebhookSecret: os.Getenv("POLAR_WEBHOOK_SECRET"),
	}
}

func AppURL() string {
	if Dev {
		return "http://localhost:3000"
	}
	return "https://afastchat.com"
}

func GetEnvOr(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
