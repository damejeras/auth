package main

import (
	"github.com/damejeras/auth/internal/app"
	"github.com/damejeras/auth/pkg/grace"
	"log"
	"net"
	"net/http"
	"os"
)

var logger = initLogger()

func main() {
	config, err := initConfig()
	if err != nil {
		log.Printf("init config: %v", err)

		os.Exit(1)
	}

	oauth2Server, err := initOauth2HTTP(config, logger)
	if err != nil {
		log.Printf("init oauth2 http: %v", err)

		os.Exit(1)
	}

	adminServer, err := initAdminHTTP(config, logger)
	if err != nil {
		log.Printf("init admin http: %v", err)

		os.Exit(1)
	}

	run(config, oauth2Server, adminServer)
}

func run(config *app.Config, oauth2, admin *http.Server) {
	ctx, cancel := grace.NewAppContext()
	defer cancel()

	adminListener, err := net.Listen("tcp", config.AdminConfig.Port)
	if err != nil {
		log.Printf("create rpc server listener: %v", err)

		os.Exit(1)
	}

	go func() {
		if err := grace.Serve(ctx, admin, adminListener); err != nil {
			log.Printf("serve rpc server: %v", err)

			cancel()
		}
	}()

	oauth2Listener, err := net.Listen("tcp", config.Oauth2Config.Port)
	if err != nil {
		log.Printf("create oauth2 server listener: %v", err)

		os.Exit(1)
	}

	if err := grace.Serve(ctx, oauth2, oauth2Listener); err != nil {
		log.Printf("serve oauth2 server: %v", err)

		os.Exit(1)
	}
}
