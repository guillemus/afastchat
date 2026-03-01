package main

import (
	"datastar-chat/logger"
	"datastar-chat/server"
)

func main() {
	logger.Setup()
	srv := server.New("1", true)
	srv.Start()
}
