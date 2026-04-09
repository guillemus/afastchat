package main

import (
	"afastchat.com/logger"
	"afastchat.com/server"
)

func main() {
	logger.Setup()
	server.StartServer()
}
