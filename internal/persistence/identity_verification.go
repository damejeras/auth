package persistence

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/damejeras/auth/internal/identity"
	"github.com/pkg/errors"
)

const tableNameIdentityVerification = "oauth_identity_verification"

type basicVerification struct {
	ChallengeID, LoginVerifier, RequestID string
	Data                                  []byte
}

type identityVerificationRepository struct {
	db *dynamodb.DynamoDB
}

func NewIdentityVerificationRepository(db *dynamodb.DynamoDB) (identity.VerificationRepository, error) {
	if err := migrateVerificationTable(db); err != nil {
		return nil, errors.Wrap(err, "run table migration")
	}

	return &identityVerificationRepository{db: db}, nil
}

func (r *identityVerificationRepository) Store(ctx context.Context, verification *identity.Verification) error {
	data, err := json.Marshal(verification.Data)
	if err != nil {
		return errors.Wrap(err, "marshal verification data")
	}

	_, err = r.db.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableNameIdentityVerification),
		Item: map[string]*dynamodb.AttributeValue{
			"ChallengeID":   {S: aws.String(verification.ChallengeID)},
			"LoginVerifier": {S: aws.String(verification.LoginVerifier)},
			"RequestID":     {S: aws.String(verification.RequestID)},
			"Data":          {B: data},
		},
	})

	return err
}

func (r *identityVerificationRepository) RetrieveByID(ctx context.Context, challengeID string) (*identity.Verification, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableNameIdentityVerification),
		Key:       map[string]*dynamodb.AttributeValue{"ChallengeID": {S: aws.String(challengeID)}},
	}

	result, err := r.db.GetItemWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(result.Item) == 0 {
		return nil, nil
	}

	var basic basicVerification
	if err := dynamodbattribute.UnmarshalMap(result.Item, &basic); err != nil {
		return nil, err
	}

	var verificationData identity.Data
	if err := json.Unmarshal(basic.Data, &verificationData); err != nil {
		return nil, err
	}

	return &identity.Verification{
		ChallengeID:   basic.ChallengeID,
		LoginVerifier: basic.LoginVerifier,
		RequestID:     basic.RequestID,
		Data:          verificationData,
	}, nil
}

func migrateVerificationTable(db *dynamodb.DynamoDB) error {
	tables, err := db.ListTables(nil)
	if err != nil {
		return err
	}

	for _, table := range tables.TableNames {
		if *table == tableNameIdentityVerification {
			return nil
		}
	}

	_, err = db.CreateTable(&dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{AttributeName: aws.String("ChallengeID"), AttributeType: aws.String("S")},
			{AttributeName: aws.String("LoginVerifier"), AttributeType: aws.String("S")},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{AttributeName: aws.String("ChallengeID"), KeyType: aws.String("HASH")},
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("LoginVerifierIndex"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{AttributeName: aws.String("LoginVerifier"), KeyType: aws.String("HASH")},
				},
				Projection: &dynamodb.Projection{
					NonKeyAttributes: []*string{aws.String("LoginVerifier")},
					ProjectionType:   aws.String("INCLUDE"),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(10),
				},
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String(tableNameIdentityVerification),
	})

	return err
}
