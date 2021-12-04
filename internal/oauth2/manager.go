package oauth2

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/damejeras/auth/pkg/dynamo"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/store"
)

func NewManager(dbClient *dynamodb.DynamoDB, clientStore *store.ClientStore) *manage.Manager {
	manager := manage.NewDefaultManager()
	manager.MustTokenStorage(dynamo.NewTokenStore(dbClient))
	manager.MapClientStorage(clientStore)

	return manager
}
