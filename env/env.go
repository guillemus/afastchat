package env

import (
	"fmt"
	"os"
	"time"

	"afastchat.com/validate"
	"github.com/joho/godotenv"
)

type Config struct {
	DBPath       string `validate:"required"`
	OpenAIAPIKey string `validate:"required"`
	Port         string `validate:"required"`
	AssetVersion string `validate:"required"`
	Dev          bool   `validate:"required"`
}

var Env = parseEnv()

func parseEnv() *Config {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		panic(fmt.Sprintf("load .env file: %v", err))
	}

	cfg := &Config{
		DBPath:       getEnvOr("DB_PATH", "./tmp/app.db"),
		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
		Port:         getEnvOr("PORT", "3000"),
		AssetVersion: fmt.Sprintf("%d", time.Now().Unix()),
		Dev:          os.Getenv("DEV") == "true",
	}

	if err := validate.V.Struct(cfg); err != nil {
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
