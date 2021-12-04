package persistence

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/damejeras/auth/internal/identity"
	"github.com/pkg/errors"
)

const tableIdentityChallenge = "oauth_identity_challenge"

type identityChallengeRepository struct {
	db *dynamodb.DynamoDB
}

func NewIdentityChallengeRepository(db *dynamodb.DynamoDB) (identity.ChallengeRepository, error) {
	if err := migrateChallengeTable(db); err != nil {
		return nil, errors.Wrap(err, "run table migration")
	}

	return &identityChallengeRepository{db: db}, nil
}

func (r *identityChallengeRepository) Store(ctx context.Context, challenge *identity.Challenge) error {
	_, err := r.db.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableIdentityChallenge),
		Item: map[string]*dynamodb.AttributeValue{
			"ID":         {S: aws.String(challenge.ID)},
			"RequestID":  {S: aws.String(challenge.RequestID)},
			"RequestURL": {S: aws.String(challenge.RequestURL)},
		},
	})

	return err
}

func (r *identityChallengeRepository) RetrievePendingByID(ctx context.Context, id string) (*identity.Challenge, error) {
	result, err := r.db.TransactGetItemsWithContext(ctx, &dynamodb.TransactGetItemsInput{
		TransactItems: []*dynamodb.TransactGetItem{{
			Get: &dynamodb.Get{
				Key:       map[string]*dynamodb.AttributeValue{"ID": {S: aws.String(id)}},
				TableName: aws.String(tableIdentityChallenge),
			},
		}, {
			Get: &dynamodb.Get{
				Key:       map[string]*dynamodb.AttributeValue{"ChallengeID": {S: aws.String(id)}},
				TableName: aws.String(tableNameIdentityVerification),
			},
		}},
	})
	if err != nil {
		return nil, err
	}

	if len(result.Responses) != 2 {
		return nil, errors.Errorf("expected 2 entries, got %r", len(result.Responses))
	}

	var challenge identity.Challenge
	err = dynamodbattribute.UnmarshalMap(result.Responses[0].Item, &challenge)
	if err != nil {
		return nil, err
	}

	if len(result.Responses[1].Item) == 0 {
		return &challenge, nil
	}

	return nil, nil
}

func migrateChallengeTable(db *dynamodb.DynamoDB) error {
	tables, err := db.ListTables(nil)
	if err != nil {
		return err
	}

	for _, table := range tables.TableNames {
		if *table == tableIdentityChallenge {
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
		TableName: aws.String(tableIdentityChallenge),
	})

	return err
}
