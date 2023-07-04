package database

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// an item that will be voted on
type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func createItemTable(client *dynamodb.Client) (Table, error) {
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
		TableName: aws.String("Items"),
	}
	_, err := client.CreateTable(context.TODO(), input)
	if err != nil {
		return Table{}, err
	}
	return Table{Name: "Items", Client: client}, nil
}

func (t Table) PutItem(item Item) error {
	input := &dynamodb.PutItemInput{
		Item: map[string]types.AttributeValue{
			"ID":          &types.AttributeValueMemberS{Value: item.ID},
			"Name":        &types.AttributeValueMemberS{Value: item.Name},
			"Description": &types.AttributeValueMemberS{Value: item.Description},
		},
		TableName: aws.String(t.Name),
	}
	_, err := t.Client.PutItem(context.TODO(), input)
	return err
}

func (t Table) GetItem(id string) (Item, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{Value: id},
		},
		TableName: aws.String(t.Name),
	}
	output, err := t.Client.GetItem(context.Background(), input)
	if err != nil {
		return Item{}, err
	}
	return Item{
		ID:          output.Item["ID"].(*types.AttributeValueMemberS).Value,
		Name:        output.Item["Name"].(*types.AttributeValueMemberS).Value,
		Description: output.Item["Description"].(*types.AttributeValueMemberS).Value,
	}, nil
}
