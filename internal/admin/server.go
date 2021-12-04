package admin

import (
	"github.com/damejeras/auth/api"
	"github.com/pacedotdev/oto/otohttp"
)

func NewServer(identityService api.IdentityService) *otohttp.Server {
	rpcServer := otohttp.NewServer()
	rpcServer.Basepath = "/api/"

	api.RegisterIdentityService(rpcServer, identityService)

	return rpcServer
}
