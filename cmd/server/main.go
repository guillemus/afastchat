package main

import "t3chat-datastar/server"

func main() {
	srv := server.New("1", true)
	srv.Start()
}
