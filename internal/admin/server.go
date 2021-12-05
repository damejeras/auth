package admin

import (
	"github.com/damejeras/auth/api"
	"github.com/pacedotdev/oto/otohttp"
)

func NewServer(identityService api.IdentityService, consentService api.ConsentService) *otohttp.Server {
	rpcServer := otohttp.NewServer()
	rpcServer.Basepath = "/api/"

	api.RegisterIdentityService(rpcServer, identityService)
	api.RegisterConsentService(rpcServer, consentService)

	return rpcServer
}
