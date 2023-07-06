package database

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type UserTable Table

// a user who can vote on items
// TODO: add fields for credentials
type User struct {
	Name string `json:"name"`
}

func CreateUserTable(client *dynamodb.Client) (UserTable, error) {
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
		TableName:   aws.String("Users"),
		BillingMode: types.BillingModePayPerRequest,
	}
	_, err := client.CreateTable(context.TODO(), input)
	if err != nil {
		return UserTable{}, err
	}
	return UserTable{Name: "Users", Client: client}, nil
}

func (t UserTable) PutUser(user User) error {
	input := &dynamodb.PutItemInput{
		Item: map[string]types.AttributeValue{
			"Name": &types.AttributeValueMemberS{Value: user.Name},
		},
		TableName: aws.String(t.Name),
	}
	_, err := t.Client.PutItem(context.TODO(), input)
	return err
}

func (t UserTable) GetUser(name string) (User, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"Name": &types.AttributeValueMemberS{Value: name},
		},
		TableName: aws.String(t.Name),
	}
	output, err := t.Client.GetItem(context.Background(), input)
	if err != nil {
		return User{}, err
	}
	if len(output.Item) == 0 {
		return User{}, fmt.Errorf("No user found with name %s", name)
	}
	return User{
		Name: output.Item["Name"].(*types.AttributeValueMemberS).Value,
	}, nil
}

func (t UserTable) DeleteUser(name string) error {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]types.AttributeValue{
			"Name": &types.AttributeValueMemberS{Value: name},
		},
		TableName: aws.String(t.Name),
	}
	_, err := t.Client.DeleteItem(context.Background(), input)
	return err
}

func (t UserTable) AllUsers() ([]User, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(t.Name),
	}
	output, err := t.Client.Scan(context.Background(), input)
	if err != nil {
		return nil, err
	}
	users := []User{}
	for _, item := range output.Items {
		users = append(users, User{
			Name: item["Name"].(*types.AttributeValueMemberS).Value,
		})
	}
	return users, nil
}
