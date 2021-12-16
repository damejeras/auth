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

const tableIdentityChallenge = "oauth_identity_challenge"

type challengeRepresentation struct {
	ID, ClientID, Verifier       string
	ChallengeIdentity, Footprint []byte
	CreatedAt, UpdatedAt         int64
}

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
	identityBytes, err := json.Marshal(challenge.Identity)
	if err != nil {
		return errors.Wrap(err, "marshal authorization")
	}

	footprintBytes, err := json.Marshal(challenge.Footprint)
	if err != nil {
		return errors.Wrap(err, "marshal footprint bytes")
	}

	_, err = r.db.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableIdentityChallenge),
		Item: map[string]*dynamodb.AttributeValue{
			"ID":                {S: aws.String(challenge.ID)},
			"ClientID":          {S: aws.String(challenge.ClientID)},
			"Verifier":          {S: aws.String(challenge.Verifier)},
			"ChallengeIdentity": {B: identityBytes},
			"Footprint":         {B: footprintBytes},
			"CreatedAt":         {N: aws.String(strconv.Itoa(int(time.Now().Unix())))},
			"UpdatedAt":         {N: aws.String(strconv.Itoa(0))},
		},
	})

	return errors.Wrap(err, "execute query")
}

func (r *identityChallengeRepository) UpdateWithAuthorization(ctx context.Context, challenge *identity.Challenge) error {
	authorizationBytes, err := json.Marshal(challenge.Identity)
	if err != nil {
		return errors.Wrap(err, "marshal authorization")
	}

	_, err = r.db.UpdateItemWithContext(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableIdentityChallenge),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {S: aws.String(challenge.ID)},
		},
		UpdateExpression: aws.String("SET ChallengeIdentity = :bytes, UpdatedAt = :timestamp"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":bytes":     {B: authorizationBytes},
			":timestamp": {N: aws.String(strconv.Itoa(int(time.Now().Unix())))},
		},
	})

	return errors.Wrap(err, "execute query")
}

func (r *identityChallengeRepository) Delete(ctx context.Context, challenge *identity.Challenge) error {
	_, err := r.db.DeleteItemWithContext(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(tableIdentityChallenge),
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {S: aws.String(challenge.ID)},
		},
	})

	return errors.Wrap(err, "execute query")
}

func (r *identityChallengeRepository) FindByID(ctx context.Context, id string) (*identity.Challenge, error) {
	result, err := r.db.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableIdentityChallenge),
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

	var representation challengeRepresentation
	if err := dynamodbattribute.UnmarshalMap(result.Item, &representation); err != nil {
		return nil, errors.Wrap(err, "unmarshal query result")
	}

	var authorization identity.Identity
	var footprint integrity.Footprint
	if err := json.Unmarshal(representation.ChallengeIdentity, &authorization); err != nil {
		return nil, errors.Wrap(err, "unmarshal authorization")
	}

	if err := json.Unmarshal(representation.Footprint, &footprint); err != nil {
		return nil, errors.Wrap(err, "unmarshal footprint")
	}

	return &identity.Challenge{
		ID:        representation.ID,
		ClientID:  representation.ClientID,
		Verifier:  representation.Verifier,
		Identity:  &authorization,
		Footprint: &footprint,
		CreatedAt: time.Unix(representation.CreatedAt, 0),
		UpdatedAt: time.Unix(representation.UpdatedAt, 0),
	}, nil
}

func (r *identityChallengeRepository) FindByVerifier(ctx context.Context, verifier string) (*identity.Challenge, error) {
	input := &dynamodb.QueryInput{
		IndexName: aws.String("LoginVerifierIndex"),
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
		TableName: aws.String(tableIdentityChallenge),
	}

	result, err := r.db.QueryWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, errors.Wrap(err, "execute query")
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var representation challengeRepresentation
	if err := dynamodbattribute.UnmarshalMap(result.Items[0], &representation); err != nil {
		return nil, errors.Wrap(err, "unmarshal query result")
	}

	var authorization identity.Identity
	var footprint integrity.Footprint
	if err := json.Unmarshal(representation.ChallengeIdentity, &authorization); err != nil {
		return nil, errors.Wrap(err, "unmarshal authorization")
	}

	if err := json.Unmarshal(representation.Footprint, &footprint); err != nil {
		return nil, errors.Wrap(err, "unmarshal footprint")
	}

	return &identity.Challenge{
		ID:        representation.ID,
		ClientID:  representation.ClientID,
		Verifier:  representation.Verifier,
		Identity:  &authorization,
		Footprint: &footprint,
		CreatedAt: time.Unix(representation.CreatedAt, 0),
		UpdatedAt: time.Unix(representation.UpdatedAt, 0),
	}, nil
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
			{AttributeName: aws.String("Verifier"), AttributeType: aws.String("S")},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{AttributeName: aws.String("ID"), KeyType: aws.String("HASH")},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(10),
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("LoginVerifierIndex"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{AttributeName: aws.String("Verifier"), KeyType: aws.String("HASH")},
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
		TableName: aws.String(tableIdentityChallenge),
	})

	return err
}
