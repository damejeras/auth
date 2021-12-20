package oauth2

import (
	"github.com/damejeras/auth/internal/app"
	"github.com/go-oauth2/oauth2/v4/server"
	"log"
	"net/http"
)

func NewHTTPServer(server *server.Server) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/authorize", app.RequestMiddleware(loggingMiddleware(server.HandleAuthorizeRequest)))
	mux.Handle("/token", app.RequestMiddleware(loggingMiddleware(server.HandleTokenRequest)))

	return &http.Server{
		Handler: mux,
	}
}

func loggingMiddleware(handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			log.Println(err)
		}
	}
}
