package persistence

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/damejeras/auth/internal/app"
	"log"
)

func NewDynamoDBClient(cfg *app.Config) *dynamodb.DynamoDB {
	awsConfig := aws.NewConfig().
		WithMaxRetries(cfg.AWSConfig.MaxRetries).
		WithRegion(cfg.AWSConfig.Region).
		WithEndpoint(cfg.AWSConfig.Endpoint).
		WithCredentials(credentials.NewStaticCredentials(cfg.AWSConfig.ID, cfg.AWSConfig.Secret, ""))

	awsSession, err := session.NewSession(awsConfig)
	if err != nil {
		log.Fatalf("create AWS session: %v", err)
	}

	return dynamodb.New(awsSession)
}
