package database

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Table struct {
	Name   string
	Client *dynamodb.Client
}

type Database struct {
	Items      ItemTable
	Users      UserTable
	UserScores UserScoreTable
}

func GetClient(region string) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return dynamodb.NewFromConfig(cfg), nil
}

func ListTables(client *dynamodb.Client) ([]string, error) {
	response, err := client.ListTables(context.TODO(), &dynamodb.ListTablesInput{})
	if err != nil {
		return nil, err
	}
	return response.TableNames, nil
}

func GetDatabase(client *dynamodb.Client) (Database, error) {
	currentTables, err := ListTables(client)
	if err != nil {
		return Database{}, err
	}
	var items ItemTable
	if !contains(currentTables, "Items") {
		items, err = CreateItemTable(client)
		if err != nil {
			return Database{}, err
		}
	} else {
		items = ItemTable{Name: "Items", Client: client}
	}
	var users UserTable
	if !contains(currentTables, "Users") {
		users, err = CreateUserTable(client)
		if err != nil {
			return Database{}, err
		}
	} else {
		users = UserTable{Name: "Users", Client: client}
	}
	var userScores UserScoreTable
	if !contains(currentTables, "UserScores") {
		userScores, err = CreateUserScoreTable(client)
		if err != nil {
			return Database{}, err
		}
	} else {
		userScores = UserScoreTable{Name: "UserScores", Client: client}
	}
	return Database{
		Items:      items,
		Users:      users,
		UserScores: userScores,
	}, nil
}
