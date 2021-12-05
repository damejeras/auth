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

const tableConsentGrant = "oauth_consent_grant"

type consentGrantRepository struct {
	db *dynamodb.DynamoDB
}

type basicGrant struct {
	ID, ClientID, SubjectID, ChallengeID, RequestID, OriginURL string
	Scope                                                      []byte
}

func NewConsentGrantRepository(db *dynamodb.DynamoDB) (consent.GrantRepository, error) {
	if err := migrateConsentGrant(db); err != nil {
		return nil, errors.Wrap(err, "run table migration")
	}

	return &consentGrantRepository{db: db}, nil
}

func (r *consentGrantRepository) Store(ctx context.Context, challenge *consent.Grant) error {
	data, err := json.Marshal(challenge.Scope)
	if err != nil {
		return errors.Wrap(err, "marshal scope")
	}

	_, err = r.db.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableConsentGrant),
		Item: map[string]*dynamodb.AttributeValue{
			"ID":          {S: aws.String(challenge.ID)},
			"ClientID":    {S: aws.String(challenge.ClientID)},
			"SubjectID":   {S: aws.String(challenge.SubjectID)},
			"ChallengeID": {S: aws.String(challenge.ChallengeID)},
			"RequestID":   {S: aws.String(challenge.RequestID)},
			"OriginURL":   {S: aws.String(challenge.OriginURL)},
			"Scope":       {B: data},
		},
	})

	return err
}

func (r *consentGrantRepository) FindByID(ctx context.Context, id string) (*consent.Grant, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableConsentGrant),
		Key:       map[string]*dynamodb.AttributeValue{"ID": {S: aws.String(id)}},
	}

	result, err := r.db.GetItemWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(result.Item) == 0 {
		return nil, nil
	}

	var basic basicGrant
	if err := dynamodbattribute.UnmarshalMap(result.Item, &basic); err != nil {
		return nil, err
	}

	var data []string
	if err := json.Unmarshal(basic.Scope, &data); err != nil {
		return nil, err
	}

	return &consent.Grant{
		ID:          basic.ID,
		ClientID:    basic.ClientID,
		SubjectID:   basic.SubjectID,
		ChallengeID: basic.ChallengeID,
		RequestID:   basic.RequestID,
		OriginURL:   basic.OriginURL,
		Scope:       data,
	}, nil
}

func (r *consentGrantRepository) FindByClientAndSubject(ctx context.Context, client, subject string) (*consent.Grant, error) {
	input := &dynamodb.QueryInput{
		IndexName: aws.String("ClientSubjectIndex"),
		KeyConditions: map[string]*dynamodb.Condition{
			"ClientID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(client),
					},
				},
			},
			"SubjectID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(subject),
					},
				},
			},
		},
		TableName: aws.String(tableConsentGrant),
	}

	result, err := r.db.QueryWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var basic basicGrant
	if err := dynamodbattribute.UnmarshalMap(result.Items[0], &basic); err != nil {
		return nil, err
	}

	var data []string
	if err := json.Unmarshal(basic.Scope, &data); err != nil {
		return nil, err
	}

	return &consent.Grant{
		ID:          basic.ID,
		ClientID:    basic.ClientID,
		SubjectID:   basic.SubjectID,
		ChallengeID: basic.ChallengeID,
		RequestID:   basic.RequestID,
		OriginURL:   basic.OriginURL,
		Scope:       data,
	}, nil
}

func migrateConsentGrant(db *dynamodb.DynamoDB) error {
	tables, err := db.ListTables(nil)
	if err != nil {
		return err
	}

	for _, table := range tables.TableNames {
		if *table == tableConsentGrant {
			return nil
		}
	}

	_, err = db.CreateTable(&dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{AttributeName: aws.String("ID"), AttributeType: aws.String("S")},
			{AttributeName: aws.String("ClientID"), AttributeType: aws.String("S")},
			{AttributeName: aws.String("SubjectID"), AttributeType: aws.String("S")},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{AttributeName: aws.String("ID"), KeyType: aws.String("HASH")},
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("ClientSubjectIndex"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{AttributeName: aws.String("ClientID"), KeyType: aws.String("HASH")},
					{AttributeName: aws.String("SubjectID"), KeyType: aws.String("RANGE")},
				},
				Projection: &dynamodb.Projection{
					NonKeyAttributes: []*string{aws.String("ClientID"), aws.String("SubjectID")},
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
		TableName: aws.String(tableConsentGrant),
	})

	return err
}
