package env

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Env struct {
	DBPath           string
	OpenAIAPIKey     string
	OpenRouterAPIKey string
	ExaAPIKey        string
	Port             string
	AssetVersion     string
	Dev              bool
}

func ParseEnv() Env {
	if err := godotenv.Load(); err != nil {
		panic(fmt.Sprintf("load .env file: %v", err))
	}

	env := Env{
		DBPath:           getEnvOr("DB_PATH", "./tmp/app.db"),
		OpenAIAPIKey:     os.Getenv("OPENAI_API_KEY"),
		OpenRouterAPIKey: os.Getenv("OPENROUTER_API_KEY"),
		ExaAPIKey:        os.Getenv("EXA_AI_API_KEY"),
		Port:             getEnvOr("PORT", "3000"),
		AssetVersion:     fmt.Sprintf("%d", time.Now().Unix()),
		Dev:              os.Getenv("DEV") == "true",
	}

	return env
}

func getEnvOr(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
