package persistence

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/damejeras/auth/internal/consent"
	"github.com/pkg/errors"
)

const tableConsentChallenge = "oauth_consent_challenge"

type consentChallengeRepository struct {
	db *dynamodb.DynamoDB
}

type basicConsentChallenge struct {
	ID, ClientID, SubjectID, RequestID, RequestURL string
	Data                                           []byte
}

func NewConsentChallengeRepository(db *dynamodb.DynamoDB) (consent.ChallengeRepository, error) {
	if err := migrateConsentChallengeTable(db); err != nil {
		return nil, errors.Wrap(err, "run table migration")
	}

	return &consentChallengeRepository{db: db}, nil
}

func (r *consentChallengeRepository) Store(ctx context.Context, challenge *consent.Challenge) error {
	data, err := json.Marshal(challenge.Data)
	if err != nil {
		return errors.Wrap(err, "marshal data")
	}

	_, err = r.db.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableConsentChallenge),
		Item: map[string]*dynamodb.AttributeValue{
			"ID":        {S: aws.String(challenge.ID)},
			"ClientID":  {S: aws.String(challenge.ClientID)},
			"SubjectID": {S: aws.String(challenge.SubjectID)},
			"RequestID": {S: aws.String(challenge.RequestID)},
			"OriginURL": {S: aws.String(challenge.OriginURL)},
			"Data":      {B: data},
		},
	})

	return err
}

func (r *consentChallengeRepository) FindByID(ctx context.Context, id string) (*consent.Challenge, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableConsentChallenge),
		Key:       map[string]*dynamodb.AttributeValue{"ID": {S: aws.String(id)}},
	}

	result, err := r.db.GetItemWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(result.Item) == 0 {
		return nil, nil
	}

	var basic basicConsentChallenge
	if err := dynamodbattribute.UnmarshalMap(result.Item, &basic); err != nil {
		return nil, err
	}

	var data consent.ChallengeData
	if err := json.Unmarshal(basic.Data, &data); err != nil {
		return nil, err
	}

	return &consent.Challenge{
		ID:        id,
		ClientID:  basic.ClientID,
		SubjectID: basic.SubjectID,
		RequestID: basic.RequestID,
		OriginURL: basic.RequestURL,
		Data:      data,
	}, nil
}

func migrateConsentChallengeTable(db *dynamodb.DynamoDB) error {
	tables, err := db.ListTables(nil)
	if err != nil {
		return err
	}

	for _, table := range tables.TableNames {
		if *table == tableConsentChallenge {
			return nil
		}
	}

	_, err = db.CreateTable(&dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{AttributeName: aws.String("ID"), AttributeType: aws.String("S")},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{AttributeName: aws.String("ID"), KeyType: aws.String("HASH")},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String(tableConsentChallenge),
	})

	return err
}
