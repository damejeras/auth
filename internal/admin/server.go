package admin

import (
	"github.com/damejeras/auth/api"
	"github.com/pacedotdev/oto/otohttp"
	"net/http"
)

func NewHTTPServer(identityService api.IdentityService, consentService api.ConsentService) *http.Server {
	rpcServer := otohttp.NewServer()
	rpcServer.Basepath = "/api/"

	api.RegisterIdentityService(rpcServer, identityService)
	api.RegisterConsentService(rpcServer, consentService)

	return &http.Server{
		Handler: rpcServer,
	}
}
