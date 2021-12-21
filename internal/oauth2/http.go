package oauth2

import (
	"github.com/damejeras/auth/internal/app"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/rs/zerolog"
	"net/http"
)

func NewHTTPServer(server *server.Server, logger *zerolog.Logger) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/authorize", app.RequestMiddleware(requestLogger(logger)(server.HandleAuthorizeRequest)))
	mux.Handle("/token", app.RequestMiddleware(requestLogger(logger)(server.HandleTokenRequest)))

	return &http.Server{
		Handler: mux,
	}
}

func requestLogger(logger *zerolog.Logger) func(handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if err := handler(w, r); err != nil {
				logger.Error().Err(err).Msg("serve request")
			}
		}
	}
}
