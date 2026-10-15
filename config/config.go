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
	OpenRouterTokenKey string
	GoogleClientID     string
	GoogleClientSecret string
	R2Endpoint         string
	R2AccessKeyID      string
	R2SecretAccessKey  string
	R2Bucket           string
}

func ParseConfig() Config {
	_ = godotenv.Load()

	Dev = os.Getenv("DEV") == "true"

	return Config{
		LogLevel:           GetEnvOr("LOG_LEVEL", "info"),
		DBPath:             GetEnvOr("DB_PATH", "./tmp/app.db"),
		Port:               GetEnvOr("PORT", "3000"),
		OpenRouterTokenKey: os.Getenv("OPENROUTER_TOKEN_KEY"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		R2Endpoint:         os.Getenv("R2_ENDPOINT"),
		R2AccessKeyID:      os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey:  os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2Bucket:           os.Getenv("R2_BUCKET"),
	}
}

func AppURL() string {
	if Dev {
		return "https://localhost"
	}
	return GetEnvOr("APP_URL", "https://afastchat.com")
}

func GetEnvOr(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
