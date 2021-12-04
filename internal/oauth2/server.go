package oauth2

import (
	"github.com/damejeras/auth/internal/identity"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
)

func NewServer(manager *manage.Manager, identityManager *identity.Manager) *server.Server {
	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetAllowedGrantType(oauth2.AuthorizationCode, oauth2.ClientCredentials)
	srv.SetClientInfoHandler(server.ClientBasicHandler)
	srv.SetUserAuthorizationHandler(identityManager.UserAuthorizationHandler())

	return srv
}
