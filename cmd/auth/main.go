package main

import (
	"github.com/damejeras/auth/internal/app"
	"github.com/damejeras/auth/pkg/grace"
	"net"
	"net/http"
	"os"
)

var logger = initLogger()

func main() {
	config, err := initConfig()
	if err != nil {
		logger.Fatal().Msgf("init config: %v", err)

		os.Exit(1)
	}

	oauth2Server, err := initOauth2HTTP(config, logger)
	if err != nil {
		logger.Fatal().Msgf("init oauth2 http: %v", err)

		os.Exit(1)
	}

	adminServer, err := initAdminHTTP(config, logger)
	if err != nil {
		logger.Fatal().Msgf("init admin http: %v", err)

		os.Exit(1)
	}

	run(config, oauth2Server, adminServer)
}

func run(config *app.Config, oauth2, admin *http.Server) {
	ctx, cancel := grace.NewAppContext()
	defer cancel()

	adminListener, err := net.Listen("tcp", config.AdminConfig.Port)
	if err != nil {
		logger.Fatal().Msgf("create rpc server listener: %v", err)

		os.Exit(1)
	}

	go func() {
		logger.Info().Msgf("serving rpc server on %q", adminListener.Addr().String())
		defer logger.Info().Msgf("rpc server on %q stopped", adminListener.Addr().String())
		if err := grace.Serve(ctx, admin, adminListener); err != nil {
			logger.Fatal().Msgf("serve rpc server: %v", err)

			cancel()
		}
	}()

	oauth2Listener, err := net.Listen("tcp", config.Oauth2Config.Port)
	if err != nil {
		logger.Fatal().Msgf("create oauth2 server listener: %v", err)

		os.Exit(1)
	}

	logger.Info().Msgf("serving oauth2 server on %q", adminListener.Addr().String())
	defer logger.Info().Msgf("oauth2 server on %q stopped", adminListener.Addr().String())
	if err := grace.Serve(ctx, oauth2, oauth2Listener); err != nil {
		logger.Fatal().Msgf("serve oauth2 server: %v", err)

		os.Exit(1)
	}
}
