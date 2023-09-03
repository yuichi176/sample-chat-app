package main

import (
	"github.com/yuichi176/sample-chat-server/internal/server"
)

// https://github.com/gorilla/websocket/blob/main/examples/chat/README.md
func main() {
	server.RunServer()
}
