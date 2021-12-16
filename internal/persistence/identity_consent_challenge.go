package persistence

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/damejeras/auth/internal/identity"
	"github.com/damejeras/auth/internal/integrity"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

const tableIdentityConsentChallenge = "oauth2_identity_consent_challenge"

type consentChallengeRepresentation struct {
	ID              string
	Verifier        string
	ClientID        string
	SubjectID       string
	RequestedScopes []byte
	MissingScopes   []byte
	GrantedScopes   []byte
	Footprint       []byte
	Used            bool
	CreatedAt       int
	UpdatedAt       int
}

type consentChallengeRepository struct {
	db *dynamodb.DynamoDB
}

func NewConsentChallengeRepository(db *dynamodb.DynamoDB) (identity.ConsentChallengeRepository, error) {
	if err := migrateConsentChallengeTable(db); err != nil {
		return nil, errors.Wrap(err, "run table migration")
	}

	return &consentChallengeRepository{db: db}, nil
}

func (c *consentChallengeRepository) Store(ctx context.Context, challenge *identity.ConsentChallenge) error {
	requestedScopes, err := json.Marshal(challenge.RequestedScopes)
	if err != nil {
		return errors.Wrap(err, "marshal requested scopes")
	}

	missingScopes, err := json.Marshal(challenge.MissingScopes)
	if err != nil {
		return errors.Wrap(err, "marshal missing scopes")
	}

	grantedScopes, err := json.Marshal(challenge.GrantedScopes)
	if err != nil {
		return errors.Wrap(err, "marshal granted scopes")
	}

	footprint, err := json.Marshal(challenge.Footprint)
	if err != nil {
		return errors.Wrap(err, "marshal footprint")
	}

	_, err = c.db.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableIdentityConsentChallenge),
		Item: map[string]*dynamodb.AttributeValue{
			"ID":              {S: aws.String(challenge.ID)},
			"Verifier":        {S: aws.String(challenge.Verifier)},
			"ClientID":        {S: aws.String(challenge.ClientID)},
			"SubjectID":       {S: aws.String(challenge.SubjectID)},
			"RequestedScopes": {B: requestedScopes},
			"MissingScopes":   {B: missingScopes},
			"GrantedScopes":   {B: grantedScopes},
			"Footprint":       {B: footprint},
			"Used":            {BOOL: aws.Bool(challenge.Used)},
			"CreatedAt":       {N: aws.String(strconv.Itoa(int(time.Now().Unix())))},
			"UpdatedAt":       {N: aws.String(strconv.Itoa(0))},
		},
	})

	return errors.Wrap(err, "execute query")
}

func (c *consentChallengeRepository) UpdateWithGrantedScopes(ctx context.Context, challenge *identity.ConsentChallenge) error {
	grantedScopes, err := json.Marshal(challenge.GrantedScopes)
	if err != nil {
		return errors.Wrap(err, "marshal authorization")
	}

	_, err = c.db.UpdateItemWithContext(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableIdentityConsentChallenge),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {S: aws.String(challenge.ID)},
		},
		UpdateExpression: aws.String("SET GrantedScopes = :GrantedScopes, UpdatedAt = :UpdatedAt"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":GrantedScopes": {B: grantedScopes},
			":UpdatedAt":     {N: aws.String(strconv.Itoa(int(time.Now().Unix())))},
		},
	})

	return errors.Wrap(err, "execute query")
}

func (c *consentChallengeRepository) FindByID(ctx context.Context, id string) (*identity.ConsentChallenge, error) {
	result, err := c.db.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableIdentityConsentChallenge),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {S: aws.String(id)},
		},
	})

	if err != nil {
		return nil, errors.Wrap(err, "execute query")
	}

	if len(result.Item) == 0 {
		return nil, nil
	}

	var representation consentChallengeRepresentation
	if err := dynamodbattribute.UnmarshalMap(result.Item, &representation); err != nil {
		return nil, errors.Wrap(err, "unmarshal query result")
	}

	var requestedScopes, missingScopes, grantedScopes identity.Scopes
	if err := json.Unmarshal(representation.RequestedScopes, &requestedScopes); err != nil {
		return nil, errors.Wrap(err, "unmarshal requested scopes")
	}

	if err := json.Unmarshal(representation.MissingScopes, &missingScopes); err != nil {
		return nil, errors.Wrap(err, "unmarshal missing scopes")
	}

	if err := json.Unmarshal(representation.GrantedScopes, &grantedScopes); err != nil {
		return nil, errors.Wrap(err, "unmarshal granted scopes")
	}

	var footprint integrity.Footprint
	if err := json.Unmarshal(representation.Footprint, &footprint); err != nil {
		return nil, errors.Wrap(err, "unmarshal footprint")
	}

	return &identity.ConsentChallenge{
		ID:              representation.ID,
		Verifier:        representation.Verifier,
		ClientID:        representation.ClientID,
		SubjectID:       representation.SubjectID,
		RequestedScopes: requestedScopes,
		MissingScopes:   missingScopes,
		GrantedScopes:   grantedScopes,
		Footprint:       &footprint,
		Used:            representation.Used,
		CreatedAt:       time.Unix(int64(representation.CreatedAt), 0),
		UpdatedAt:       time.Unix(int64(representation.UpdatedAt), 0),
	}, nil
}

func (c *consentChallengeRepository) FindByVerifier(ctx context.Context, verifier string) (*identity.ConsentChallenge, error) {
	input := &dynamodb.QueryInput{
		IndexName: aws.String("ConsentVerifierIndex"),
		KeyConditions: map[string]*dynamodb.Condition{
			"Verifier": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(verifier),
					},
				},
			},
		},
		TableName: aws.String(tableIdentityConsentChallenge),
	}

	result, err := c.db.QueryWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, errors.Wrap(err, "execute query")
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var representation consentChallengeRepresentation
	if err := dynamodbattribute.UnmarshalMap(result.Items[0], &representation); err != nil {
		return nil, errors.Wrap(err, "unmarshal query result")
	}

	var requestedScopes, missingScopes, grantedScopes identity.Scopes
	if err := json.Unmarshal(representation.RequestedScopes, &requestedScopes); err != nil {
		return nil, errors.Wrap(err, "unmarshal requested scopes")
	}

	if err := json.Unmarshal(representation.MissingScopes, &missingScopes); err != nil {
		return nil, errors.Wrap(err, "unmarshal missing scopes")
	}

	if err := json.Unmarshal(representation.GrantedScopes, &grantedScopes); err != nil {
		return nil, errors.Wrap(err, "unmarshal granted scopes")
	}

	var footprint integrity.Footprint
	if err := json.Unmarshal(representation.Footprint, &footprint); err != nil {
		return nil, errors.Wrap(err, "unmarshal footprint")
	}

	return &identity.ConsentChallenge{
		ID:              representation.ID,
		Verifier:        representation.Verifier,
		ClientID:        representation.ClientID,
		SubjectID:       representation.SubjectID,
		RequestedScopes: requestedScopes,
		MissingScopes:   missingScopes,
		GrantedScopes:   grantedScopes,
		Footprint:       &footprint,
		Used:            representation.Used,
		CreatedAt:       time.Unix(int64(representation.CreatedAt), 0),
		UpdatedAt:       time.Unix(int64(representation.UpdatedAt), 0),
	}, nil
}

func (c *consentChallengeRepository) Delete(ctx context.Context, challenge *identity.ConsentChallenge) error {
	_, err := c.db.DeleteItemWithContext(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(tableIdentityConsentChallenge),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {S: aws.String(challenge.ID)},
		},
	})

	return errors.Wrap(err, "execute query")
}

func migrateConsentChallengeTable(db *dynamodb.DynamoDB) error {
	tables, err := db.ListTables(nil)
	if err != nil {
		return err
	}

	for _, table := range tables.TableNames {
		if *table == tableIdentityConsentChallenge {
			return nil
		}
	}

	_, err = db.CreateTable(&dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{AttributeName: aws.String("ID"), AttributeType: aws.String("S")},
			{AttributeName: aws.String("Verifier"), AttributeType: aws.String("S")},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{AttributeName: aws.String("ID"), KeyType: aws.String("HASH")},
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("ConsentVerifierIndex"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{AttributeName: aws.String("Verifier"), KeyType: aws.String("HASH")},
				},
				Projection: &dynamodb.Projection{
					NonKeyAttributes: []*string{aws.String("Verifier")},
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
		TableName: aws.String(tableIdentityConsentChallenge),
	})

	return err
}
