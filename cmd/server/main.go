package main

import (
	"afastchat.com/env"
	"afastchat.com/logger"
	"afastchat.com/server"
)

func main() {
	env := env.ParseEnv()
	logger.Setup(env)
	server.StartServer(env)
}
