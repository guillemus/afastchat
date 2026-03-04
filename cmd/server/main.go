package main

import (
	"datastar-chat/env"
	"datastar-chat/logger"
	"datastar-chat/server"
)

func main() {
	logger.Setup()
	srv := server.New(env.Env.DatabaseURL, env.Env.Port, env.Env.AssetVersion, env.Env.Dev)
	srv.Start()
}
