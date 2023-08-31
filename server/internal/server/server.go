package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"github.com/yuichi176/sample-chat-server/config"
	"github.com/yuichi176/sample-chat-server/internal/handlers"
	"log"
	"net"
	"net/http"
)

func RunServer() {

	listenAddr := config.GetConfig().Server.ListenAddr
	host, port, _ := net.SplitHostPort(listenAddr)

	var srv http.Server
	srv.Addr = fmt.Sprintf("%s:%s", host, port)
	srv.Handler = registerHandlers()

	if err := srv.ListenAndServe(); err != nil {
		log.Panicln("Serve Error:", err)
	}
}

func registerHandlers() http.Handler {
	n := negroni.New()

	r := mux.NewRouter()
	r.HandleFunc("/monitor/l7check", handlers.HealthCheckHandler)

	n.UseHandler(r)
	return n
}
