package client

import (
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/store"
)

func NewClientStorage() *store.ClientStore {
	// for testing purposes, should be replaced with proper implementation
	clientStore := store.NewClientStore()
	clientStore.Set("test", &models.Client{
		ID:     "test",
		Secret: "test",
		Domain: "https://oauth.tools/",
	})

	return clientStore
}
