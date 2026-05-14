package env

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Env struct {
	DBPath             string
	OpenAIAPIKey       string
	FireworksAPIKey    string
	OpenRouterAPIKey   string
	ExaAPIKey          string
	PolarAPIKey        string
	PolarWebhookSecret string
	Port               string
	AssetVersion       string
	Dev                bool
}

func ParseEnv() Env {
	if err := godotenv.Load(); err != nil {
		panic(fmt.Sprintf("load .env file: %v", err))
	}

	env := Env{
		DBPath:             getEnvOr("DB_PATH", "./tmp/app.db"),
		OpenAIAPIKey:       os.Getenv("OPENAI_API_KEY"),
		FireworksAPIKey:    os.Getenv("FIREWORKS_API_KEY"),
		OpenRouterAPIKey:   os.Getenv("OPENROUTER_API_KEY"),
		ExaAPIKey:          os.Getenv("EXA_AI_API_KEY"),
		PolarAPIKey:        os.Getenv("POLAR_ACCESS_TOKEN"),
		PolarWebhookSecret: os.Getenv("POLAR_WEBHOOK_SECRET"),
		Port:               getEnvOr("PORT", "3000"),
		AssetVersion:       fmt.Sprintf("%d", time.Now().Unix()),
		Dev:                os.Getenv("DEV") == "true",
	}

	return env
}

func AppURL(env Env) string {
	if env.Dev {
		return "http://localhost:3000"
	}
	return "https://afastchat.com"
}

func getEnvOr(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
