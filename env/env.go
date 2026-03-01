package env

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"datastar-chat/validate"
)

type Config struct {
	DatabaseURL  string `env:"DATABASE_URL" validate:"required"`
	OpenAIAPIKey string `env:"OPENAI_API_KEY" validate:"required"`
	Port         string `env:"PORT"`
	AssetVersion string `env:"RAILWAY_DEPLOYMENT_ID"`
	Dev          bool   `env:"DEV"`
}

var Env = parseEnv()

func parseEnv() *Config {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		panic(fmt.Sprintf("load .env file: %v", err))
	}

	cfg := &Config{
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
		Port:         os.Getenv("PORT"),
		AssetVersion: os.Getenv("RAILWAY_DEPLOYMENT_ID"),
		Dev:          os.Getenv("DEV") == "true",
	}
	if cfg.AssetVersion == "" {
		cfg.AssetVersion = fmt.Sprintf("%d", time.Now().Unix())
	}
	if cfg.Port == "" {
		cfg.Port = "3000"
	}

	if err := validate.V.Struct(cfg); err != nil {
		panic(fmt.Sprintf("validate env vars: %v", err))
	}

	return cfg
}
