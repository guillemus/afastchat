package main

import (
	"datastar-chat/env"
	"datastar-chat/logger"
	"datastar-chat/server"
)

func main() {
	logger.Setup()
	server.StartServer(env.Env.DatabaseURL, env.Env.Port, env.Env.AssetVersion, env.Env.Dev)
}
