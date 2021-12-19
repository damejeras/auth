package persistence

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/damejeras/auth/internal/consent"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

const tableConsent = "oauth2_consent"

type consentRepresentation struct {
	ID        string
	ClientID  string
	SubjectID string
	Scopes    []byte
	CreatedAt int64
	UpdatedAt int64
}

type consentRepository struct {
	db *dynamodb.DynamoDB
}

func NewConsentRepository(db *dynamodb.DynamoDB) (consent.Repository, error) {
	if err := migrateIdentityConsentTable(db); err != nil {
		return nil, errors.Wrap(err, "run table migration")
	}

	return &consentRepository{db: db}, nil
}

func (c *consentRepository) Store(ctx context.Context, consent *consent.Consent) error {
	scopeBytes, err := json.Marshal(consent.Scopes)
	if err != nil {
		return errors.Wrap(err, "marshal scopes")
	}

	_, err = c.db.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableConsent),
		Item: map[string]*dynamodb.AttributeValue{
			"ID":        {S: aws.String(consent.ID)},
			"ClientID":  {S: aws.String(consent.ClientID)},
			"SubjectID": {S: aws.String(consent.SubjectID)},
			"Scopes":    {B: scopeBytes},
			"CreatedAt": {N: aws.String(strconv.Itoa(int(time.Now().Unix())))},
			"UpdatedAt": {N: aws.String(strconv.Itoa(0))},
		},
	})

	return errors.Wrap(err, "execute query")
}

func (c *consentRepository) UpdateWithScopes(ctx context.Context, consent *consent.Consent) error {
	scopeBytes, err := json.Marshal(consent.Scopes)
	if err != nil {
		return errors.Wrap(err, "marshal scopes")
	}

	_, err = c.db.UpdateItemWithContext(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableConsent),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {S: aws.String(consent.ID)},
		},
		UpdateExpression: aws.String("SET Scopes = :Scopes, UpdatedAt = :UpdatedAt"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":Scopes":    {B: scopeBytes},
			":UpdatedAt": {N: aws.String(strconv.Itoa(int(time.Now().Unix())))},
		},
	})

	return errors.Wrap(err, "execute query")
}

func (c *consentRepository) FindByClientAndSubject(ctx context.Context, clientID, subjectID string) (*consent.Consent, error) {
	input := &dynamodb.QueryInput{
		IndexName: aws.String("ClientSubjectIndex"),
		KeyConditions: map[string]*dynamodb.Condition{
			"ClientID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(clientID),
					},
				},
			},
			"SubjectID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(subjectID),
					},
				},
			},
		},
		TableName: aws.String(tableConsent),
	}

	result, err := c.db.QueryWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var representation consentRepresentation
	if err := dynamodbattribute.UnmarshalMap(result.Items[0], &representation); err != nil {
		return nil, err
	}

	var scopes consent.Scopes
	if err := json.Unmarshal(representation.Scopes, &scopes); err != nil {
		return nil, err
	}

	return &consent.Consent{
		ID:        representation.ID,
		ClientID:  representation.ClientID,
		SubjectID: representation.SubjectID,
		Scopes:    scopes,
		CreatedAt: time.Unix(representation.CreatedAt, 0),
		UpdatedAt: time.Unix(representation.UpdatedAt, 0),
	}, nil
}

func migrateIdentityConsentTable(db *dynamodb.DynamoDB) error {
	tables, err := db.ListTables(nil)
	if err != nil {
		return err
	}

	for _, table := range tables.TableNames {
		if *table == tableConsent {
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
					ProjectionType: aws.String("ALL"),
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
		TableName: aws.String(tableConsent),
	})

	return err
}
