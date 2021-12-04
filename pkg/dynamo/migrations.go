package dynamo

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func (ts *tokenStore) runMigrations() error {
	tables, err := ts.dbClient.ListTables(nil)
	if err != nil {
		return err
	}

	tableMap := make(map[string]struct{})

	for _, table := range tables.TableNames {
		if table != nil {
			tableMap[*table] = struct{}{}
		}
	}

	for _, tableName := range []string{
		ts.tables.BasicCname,
		ts.tables.AccessCName,
		ts.tables.RefreshCName,
	} {
		if _, ok := tableMap[tableName]; !ok {
			if err := ts.createSingleTable(tableName); err != nil {
				return err
			}
		}
	}

	return nil
}

func (ts *tokenStore) createSingleTable(name string) error {
	_, err := ts.dbClient.CreateTable(&dynamodb.CreateTableInput{
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
		TableName: aws.String(name),
	})

	return err
}
