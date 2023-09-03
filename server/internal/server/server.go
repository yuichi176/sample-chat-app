package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"github.com/yuichi176/sample-chat-server/config"
	"github.com/yuichi176/sample-chat-server/internal/chat"
	"github.com/yuichi176/sample-chat-server/internal/handlers"
	"log"
	"net"
	"net/http"
)

func RunServer() {
	// WebSocketハブを起動
	// クライアントの登録、登録解除、メッセージのブロードキャストを非同期に処理
	hub := chat.NewHub()
	go hub.Run()

	// 設定ファイルの読み込み
	listenAddr := config.GetConfig().Server.ListenAddr
	host, port, _ := net.SplitHostPort(listenAddr)

	var srv http.Server
	srv.Addr = fmt.Sprintf("%s:%s", host, port)
	srv.Handler = registerHandlers(hub)

	if err := srv.ListenAndServe(); err != nil {
		log.Panicln("Serve Error:", err)
	}
}

func registerHandlers(hub *chat.Hub) http.Handler {
	// https://github.com/urfave/negroni
	n := negroni.Classic() // Includes some default middlewares

	// https://github.com/gorilla/mux
	r := mux.NewRouter()
	r.HandleFunc("/", handlers.ServeHome)
	r.HandleFunc("/monitor/l7check", handlers.HealthCheckHandler)
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		chat.ServeWs(hub, w, r)
	})

	n.UseHandler(r)
	return n
}
