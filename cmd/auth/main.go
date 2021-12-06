package main

import (
	"log"
	"net/http"
	"os"

	"github.com/damejeras/auth/internal/application"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/pacedotdev/oto/otohttp"
)

func main() {
	oauth2Server, err := InitializeOauth2Server()
	if err != nil {
		log.Printf("initialize oauth2 server: %v", err)

		os.Exit(1)
	}

	rpcServer, err := InitializeRPCServer()
	if err != nil {
		log.Printf("initialize rpc server: %v", err)

		os.Exit(1)
	}

	run(oauth2Server, rpcServer)
}

func run(oauth2Server *server.Server, rpcServer *otohttp.Server) {
	http.Handle("/authorize", application.ContextMiddleware(wrapOauthServerHandlers(oauth2Server.HandleAuthorizeRequest)))
	http.Handle("/token", application.ContextMiddleware(wrapOauthServerHandlers(oauth2Server.HandleTokenRequest)))

	go func() {
		if err := http.ListenAndServe(":9097", rpcServer); err != nil {
			log.Printf("serve rpc server: %v", err)
		}
	}()

	if err := http.ListenAndServe(":9096", nil); err != nil {
		log.Printf("listen and serve: %v", err)

		os.Exit(1)
	}
}

func wrapOauthServerHandlers(handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			log.Println(err)
		}
	}
}
