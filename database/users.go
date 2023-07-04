package database

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// a user
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func createUserTable(client *dynamodb.Client) (Table, error) {
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("ID"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("ID"),
				KeyType:       types.KeyTypeHash,
			},
		},
		TableName: aws.String("Users"),
	}
	_, err := client.CreateTable(context.TODO(), input)
	if err != nil {
		return Table{}, err
	}
	return Table{Name: "Users", Client: client}, nil
}
