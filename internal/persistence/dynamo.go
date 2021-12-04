package persistence

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
)

func NewDynamoDBClient() *dynamodb.DynamoDB {
	awsConfig := aws.NewConfig()
	awsConfig.Region = aws.String("eu-west-1")
	awsConfig.Endpoint = aws.String("http://localhost:8000")
	awsConfig.Credentials = credentials.NewStaticCredentials("123", "123", "123")

	awsSession, err := session.NewSession(awsConfig)
	if err != nil {
		log.Fatalf("create AWS session: %v", err)
	}

	return dynamodb.New(awsSession)
}
