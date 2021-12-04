package dynamo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
	"gopkg.in/mgo.v2/bson"
)

type tokenStore struct {
	tables   TableConfig
	dbClient *dynamodb.DynamoDB
}

func NewTokenStore(client *dynamodb.DynamoDB, options ...Option) (oauth2.TokenStore, error) {
	store := tokenStore{
		tables:   DefaultTableConfig,
		dbClient: client,
	}

	if err := store.runMigrations(); err != nil {
		return nil, err
	}

	for i := range options {
		options[i](&store)
	}

	return &store, nil
}

func (ts *tokenStore) Create(ctx context.Context, info oauth2.TokenInfo) error {
	if code := info.GetCode(); code != "" {
		err := createWithAuthorizationCode(ctx, ts, info, "")
		if err != nil {
			return err
		}

		return nil
	}

	if refresh := info.GetRefresh(); refresh != "" {
		if err := createWithRefreshToken(ctx, ts, info); err != nil {
			return err
		}
	} else {
		if err := createWithAccessToken(ctx, ts, info, ""); err != nil {
			return err
		}
	}

	return nil
}

func (ts *tokenStore) RemoveByCode(ctx context.Context, code string) error {
	input := &dynamodb.DeleteItemInput{
		Key:       map[string]*dynamodb.AttributeValue{"ID": {S: aws.String(code)}},
		TableName: aws.String(ts.tables.BasicCname),
	}

	_, err := ts.dbClient.DeleteItemWithContext(ctx, input)
	if err != nil {
		return err
	}

	return nil
}

func (ts *tokenStore) RemoveByAccess(ctx context.Context, access string) error {
	input := &dynamodb.DeleteItemInput{
		Key:       map[string]*dynamodb.AttributeValue{"ID": {S: aws.String(access)}},
		TableName: aws.String(ts.tables.AccessCName),
	}

	_, err := ts.dbClient.DeleteItemWithContext(ctx, input)
	if err != nil {
		return err
	}

	return nil
}

func (ts *tokenStore) RemoveByRefresh(ctx context.Context, refresh string) error {
	input := &dynamodb.DeleteItemInput{
		Key:       map[string]*dynamodb.AttributeValue{"ID": {S: aws.String(refresh)}},
		TableName: aws.String(ts.tables.RefreshCName),
	}

	_, err := ts.dbClient.DeleteItemWithContext(ctx, input)
	if err != nil {
		return err
	}

	return nil
}

func (ts *tokenStore) GetByCode(ctx context.Context, code string) (oauth2.TokenInfo, error) {
	return ts.getData(ctx, code)
}

func (ts *tokenStore) GetByAccess(ctx context.Context, access string) (oauth2.TokenInfo, error) {
	basicID, err := ts.getBasicID(ctx, ts.tables.AccessCName, access)
	if err != nil && basicID == "" {
		return nil, err
	}

	return ts.getData(ctx, basicID)
}

func (ts *tokenStore) GetByRefresh(ctx context.Context, refresh string) (oauth2.TokenInfo, error) {
	basicID, err := ts.getBasicID(ctx, ts.tables.RefreshCName, refresh)
	if err != nil && basicID == "" {
		return nil, err
	}

	return ts.getData(ctx, basicID)
}

type tokenData struct {
	ID        string    `json:"_id"`
	BasicID   string    `json:"BasicID"`
	ExpiredAt time.Time `json:"ExpiredAt"`
}

type basicData struct {
	ID        string    `json:"_id"`
	Data      []byte    `json:"Data"`
	ExpiredAt time.Time `json:"ExpiredAt"`
}

func createWithAuthorizationCode(ctx context.Context, tokenStorage *tokenStore, info oauth2.TokenInfo, id string) error {
	code := info.GetCode()
	if len(id) > 0 {
		code = id
	}

	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	expiredAt := info.GetCodeCreateAt().Add(info.GetCodeExpiresIn())
	rExpiredAt := expiredAt
	if refresh := info.GetRefresh(); refresh != "" {
		refreshExpiration := info.GetRefreshCreateAt().Add(info.GetRefreshExpiresIn())
		if expiredAt.Second() > refreshExpiration.Second() {
			expiredAt = refreshExpiration
		}

		rExpiredAt = refreshExpiration
	}

	exp := rExpiredAt.Format(time.RFC3339)
	params := &dynamodb.PutItemInput{
		TableName: aws.String(tokenStorage.tables.BasicCname),
		Item: map[string]*dynamodb.AttributeValue{
			"ID":        {S: aws.String(code)},
			"Data":      {B: data},
			"ExpiredAt": {S: &exp},
		},
	}

	_, err = tokenStorage.dbClient.PutItemWithContext(ctx, params)

	return err
}

func createWithAccessToken(ctx context.Context, tokenStorage *tokenStore, info oauth2.TokenInfo, id string) error {
	if len(id) == 0 {
		id = bson.NewObjectId().Hex()
	}

	err := createWithAuthorizationCode(ctx, tokenStorage, info, id)
	if err != nil {
		return err
	}

	expiredAt := info.GetAccessCreateAt().Add(info.GetAccessExpiresIn()).Format(time.RFC3339)
	accessParams := &dynamodb.PutItemInput{
		TableName: aws.String(tokenStorage.tables.AccessCName),
		Item: map[string]*dynamodb.AttributeValue{
			"ID":        {S: aws.String(info.GetAccess())},
			"BasicID":   {S: &id},
			"ExpiredAt": {S: &expiredAt},
		},
	}

	_, err = tokenStorage.dbClient.PutItem(accessParams)

	return err
}

func createWithRefreshToken(ctx context.Context, tokenStorage *tokenStore, info oauth2.TokenInfo) error {
	id := bson.NewObjectId().Hex()

	err := createWithAccessToken(ctx, tokenStorage, info, id)
	if err != nil {
		return err
	}

	expiredAt := info.GetRefreshCreateAt().Add(info.GetRefreshExpiresIn()).Format(time.RFC3339)
	refreshParams := &dynamodb.PutItemInput{
		TableName: aws.String(tokenStorage.tables.RefreshCName),
		Item: map[string]*dynamodb.AttributeValue{
			"ID":        {S: aws.String(info.GetRefresh())},
			"BasicID":   {S: &id},
			"ExpiredAt": {S: &expiredAt},
		},
	}
	_, err = tokenStorage.dbClient.PutItem(refreshParams)

	return err
}

func (ts *tokenStore) getData(ctx context.Context, basicID string) (oauth2.TokenInfo, error) {
	if len(basicID) == 0 {
		return nil, nil
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(ts.tables.BasicCname),
		Key:       map[string]*dynamodb.AttributeValue{"ID": {S: aws.String(basicID)}},
	}

	result, err := ts.dbClient.GetItemWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(result.Item) == 0 {
		return nil, nil
	}

	var b basicData
	err = dynamodbattribute.UnmarshalMap(result.Item, &b)
	if err != nil {
		return nil, err
	}

	var tm models.Token
	err = json.Unmarshal(b.Data, &tm)
	if err != nil {
		return nil, err
	}

	return &tm, nil
}

func (ts *tokenStore) getBasicID(ctx context.Context, cname, token string) (string, error) {
	input := &dynamodb.GetItemInput{
		Key:       map[string]*dynamodb.AttributeValue{"ID": {S: aws.String(token)}},
		TableName: aws.String(cname),
	}

	result, err := ts.dbClient.GetItemWithContext(ctx, input)
	if err != nil {
		return "", err
	}

	var td tokenData
	err = dynamodbattribute.UnmarshalMap(result.Item, &td)
	if err != nil {
		return "", err
	}

	return td.BasicID, nil
}
