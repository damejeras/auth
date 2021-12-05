//go:build wireinject

package main

import (
	"github.com/damejeras/auth/internal/admin"
	"github.com/damejeras/auth/internal/client"
	"github.com/damejeras/auth/internal/consent"
	"github.com/damejeras/auth/internal/identity"
	"github.com/damejeras/auth/internal/oauth2"
	"github.com/damejeras/auth/internal/persistence"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/google/wire"
	"github.com/pacedotdev/oto/otohttp"
)

func InitializeOauth2Server() (*server.Server, error) {
	wire.Build(
		oauth2.NewServer,
		oauth2.NewManager,
		client.NewClientStorage,
		identity.NewManager,
		persistence.NewDynamoDBClient,
		persistence.NewIdentityChallengeRepository,
		persistence.NewIdentityVerificationRepository,
		persistence.NewConsentChallengeRepository,
		persistence.NewConsentGrantRepository,
	)

	return nil, nil
}

func InitializeRPCServer() (*otohttp.Server, error) {
	wire.Build(
		identity.NewService,
		consent.NewService,
		persistence.NewDynamoDBClient,
		persistence.NewIdentityChallengeRepository,
		persistence.NewIdentityVerificationRepository,
		persistence.NewConsentGrantRepository,
		persistence.NewConsentChallengeRepository,
		admin.NewServer,
	)

	return nil, nil
}
