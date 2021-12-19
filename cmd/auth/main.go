package main

import (
	"github.com/damejeras/auth/internal/app"
	"github.com/damejeras/auth/internal/integrity"
	"log"
	"net/http"
	"os"

	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/pacedotdev/oto/otohttp"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Printf("load config: %v", err)

		os.Exit(1)
	}

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

	run(cfg, oauth2Server, rpcServer)
}

func run(cfg *app.Config, oauth2Server *server.Server, rpcServer *otohttp.Server) {
	http.Handle("/authorize", integrity.ContextMiddleware(wrapOauthServerHandlers(oauth2Server.HandleAuthorizeRequest)))
	http.Handle("/token", integrity.ContextMiddleware(wrapOauthServerHandlers(oauth2Server.HandleTokenRequest)))

	go func() {
		log.Printf("serving rpc server on %q", cfg.API.Port)
		if err := http.ListenAndServe(cfg.API.Port, rpcServer); err != nil {
			log.Printf("serve rpc server: %v", err)
		}
	}()

	log.Printf("serving application on %q", cfg.App.Port)
	if err := http.ListenAndServe(cfg.App.Port, nil); err != nil {
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
