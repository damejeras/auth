//go:build wireinject

package main

import (
	"github.com/damejeras/auth/internal/admin"
	"github.com/damejeras/auth/internal/app"
	"github.com/damejeras/auth/internal/client"
	"github.com/damejeras/auth/internal/consent"
	"github.com/damejeras/auth/internal/identity"
	"github.com/damejeras/auth/internal/oauth2"
	"github.com/damejeras/auth/internal/persistence"
	"github.com/google/wire"
	"github.com/kkyr/fig"
	"net/http"
)

func initConfig() (*app.Config, error) {
	var config app.Config
	if err := fig.Load(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func initOauth2HTTP(cfg *app.Config) (*http.Server, error) {
	wire.Build(
		oauth2.NewHTTPServer,
		oauth2.NewServer,
		oauth2.NewManager,
		client.NewClientStorage,
		identity.NewManager,
		persistence.NewDynamoDBClient,
		persistence.NewIdentityChallengeRepository,
		persistence.NewConsentChallengeRepository,
		persistence.NewConsentRepository,
	)

	return nil, nil
}

func initAdminHTTP(cfg *app.Config) (*http.Server, error) {
	wire.Build(
		admin.NewHTTPServer,
		identity.NewService,
		consent.NewService,
		persistence.NewDynamoDBClient,
		persistence.NewIdentityChallengeRepository,
		persistence.NewConsentChallengeRepository,
		persistence.NewConsentRepository,
	)

	return nil, nil
}
