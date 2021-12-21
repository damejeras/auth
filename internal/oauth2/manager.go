package oauth2

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/damejeras/auth/pkg/dynamo"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/pkg/errors"
)

func NewManager(dbClient *dynamodb.DynamoDB, clientStore *store.ClientStore) (*manage.Manager, error) {
	manager := manage.NewDefaultManager()
	tokenStorage, err := dynamo.NewTokenStore(dbClient)
	if err != nil {
		return nil, errors.Wrap(err, "create token storage")
	}

	manager.MapTokenStorage(tokenStorage)
	manager.MapClientStorage(clientStore)

	return manager, nil
}
