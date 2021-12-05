// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/damejeras/auth/internal/admin"
	"github.com/damejeras/auth/internal/client"
	"github.com/damejeras/auth/internal/consent"
	"github.com/damejeras/auth/internal/identity"
	"github.com/damejeras/auth/internal/oauth2"
	"github.com/damejeras/auth/internal/persistence"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/pacedotdev/oto/otohttp"
)

// Injectors from wire.go:

func InitializeOauth2Server() (*server.Server, error) {
	dynamoDB := persistence.NewDynamoDBClient()
	clientStore := client.NewClientStorage()
	manager := oauth2.NewManager(dynamoDB, clientStore)
	challengeRepository, err := persistence.NewIdentityChallengeRepository(dynamoDB)
	if err != nil {
		return nil, err
	}
	verificationRepository, err := persistence.NewIdentityVerificationRepository(dynamoDB)
	if err != nil {
		return nil, err
	}
	consentChallengeRepository, err := persistence.NewConsentChallengeRepository(dynamoDB)
	if err != nil {
		return nil, err
	}
	grantRepository, err := persistence.NewConsentGrantRepository(dynamoDB)
	if err != nil {
		return nil, err
	}
	identityManager := identity.NewManager(challengeRepository, verificationRepository, consentChallengeRepository, grantRepository)
	serverServer := oauth2.NewServer(manager, identityManager)
	return serverServer, nil
}

func InitializeRPCServer() (*otohttp.Server, error) {
	dynamoDB := persistence.NewDynamoDBClient()
	challengeRepository, err := persistence.NewIdentityChallengeRepository(dynamoDB)
	if err != nil {
		return nil, err
	}
	verificationRepository, err := persistence.NewIdentityVerificationRepository(dynamoDB)
	if err != nil {
		return nil, err
	}
	identityService := identity.NewService(challengeRepository, verificationRepository)
	consentChallengeRepository, err := persistence.NewConsentChallengeRepository(dynamoDB)
	if err != nil {
		return nil, err
	}
	grantRepository, err := persistence.NewConsentGrantRepository(dynamoDB)
	if err != nil {
		return nil, err
	}
	consentService := consent.NewService(consentChallengeRepository, grantRepository)
	otohttpServer := admin.NewServer(identityService, consentService)
	return otohttpServer, nil
}
