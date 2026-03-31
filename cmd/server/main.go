package main

import (
	"datastar-chat/logger"
	"datastar-chat/server"
)

func main() {
	logger.Setup()
	server.StartServer()
}
