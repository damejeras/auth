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
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/google/wire"
	"github.com/kkyr/fig"
	"github.com/pacedotdev/oto/otohttp"
)

var (
	config app.Config
	loaded bool
)

func InitializeOauth2Server() (*server.Server, error) {
	wire.Build(
		loadConfig,
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

func InitializeRPCServer() (*otohttp.Server, error) {
	wire.Build(
		loadConfig,
		identity.NewService,
		consent.NewService,
		persistence.NewDynamoDBClient,
		persistence.NewIdentityChallengeRepository,
		persistence.NewConsentChallengeRepository,
		persistence.NewConsentRepository,
		admin.NewServer,
	)

	return nil, nil
}

func loadConfig() (*app.Config, error) {
	if loaded {
		return &config, nil
	}

	if err := fig.Load(&config); err != nil {
		return nil, err
	}

	loaded = true

	return &config, nil
}
