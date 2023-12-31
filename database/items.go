package database

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type ItemTable Table

// an item that will be voted on
type Item struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func CreateItemTable(client *dynamodb.Client) (ItemTable, error) {
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("Name"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("Name"),
				KeyType:       types.KeyTypeHash,
			},
		},
		TableName:   aws.String("Items"),
		BillingMode: types.BillingModePayPerRequest,
	}
	_, err := client.CreateTable(context.TODO(), input)
	if err != nil {
		return ItemTable{}, err
	}
	return ItemTable{Name: "Items", Client: client}, nil
}

func (t ItemTable) PutItem(item Item) error {
	input := &dynamodb.PutItemInput{
		Item: map[string]types.AttributeValue{
			"Name":        &types.AttributeValueMemberS{Value: item.Name},
			"Description": &types.AttributeValueMemberS{Value: item.Description},
		},
		TableName: aws.String(t.Name),
	}
	_, err := t.Client.PutItem(context.TODO(), input)
	return err
}

func (t ItemTable) GetItem(name string) (Item, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"Name": &types.AttributeValueMemberS{Value: name},
		},
		TableName: aws.String(t.Name),
	}
	output, err := t.Client.GetItem(context.TODO(), input)
	if err != nil {
		return Item{}, err
	}
	if output.Item == nil {
		return Item{}, fmt.Errorf("no item found with name %s", name)
	}
	return Item{
		Name:        output.Item["Name"].(*types.AttributeValueMemberS).Value,
		Description: output.Item["Description"].(*types.AttributeValueMemberS).Value,
	}, nil
}

func (t ItemTable) DeleteItem(name string) error {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]types.AttributeValue{
			"Name": &types.AttributeValueMemberS{Value: name},
		},
		TableName: aws.String(t.Name),
	}
	_, err := t.Client.DeleteItem(context.TODO(), input)
	return err
}

func (t ItemTable) AllItems() ([]Item, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(t.Name),
	}
	output, err := t.Client.Scan(context.TODO(), input)
	if err != nil {
		return nil, err
	}
	items := make([]Item, len(output.Items))
	for i, item := range output.Items {
		items[i] = Item{
			Name:        item["Name"].(*types.AttributeValueMemberS).Value,
			Description: item["Description"].(*types.AttributeValueMemberS).Value,
		}
	}
	return items, nil
}
