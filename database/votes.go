package database

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type UserScoreTable Table

// a vote on an item
type UserScore struct {
	ItemName string `json:"item_name"`
	UserName string `json:"user_name"`
	Rating   int    `json:"rating"`
}

func CreateUserScoreTable(client *dynamodb.Client) (UserScoreTable, error) {
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("ItemName"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("UserName"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("ItemName"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("UserName"),
				KeyType:       types.KeyTypeRange,
			},
		},
		TableName:   aws.String("UserScores"),
		BillingMode: types.BillingModePayPerRequest,
	}
	_, err := client.CreateTable(context.TODO(), input)
	if err != nil {
		return UserScoreTable{}, err
	}
	return UserScoreTable{Name: "UserScores", Client: client}, nil
}

func (t UserScoreTable) PutUserScore(u UserScore) error {
	input := &dynamodb.PutItemInput{
		Item: map[string]types.AttributeValue{
			"ItemName": &types.AttributeValueMemberS{Value: u.ItemName},
			"UserName": &types.AttributeValueMemberS{Value: u.UserName},
			"Rating":   &types.AttributeValueMemberN{Value: strconv.Itoa(u.Rating)},
		},
		TableName: aws.String(t.Name),
	}
	_, err := t.Client.PutItem(context.TODO(), input)
	return err
}

func (t UserScoreTable) UpdateUserScore(u UserScore) error {
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":rating": &types.AttributeValueMemberN{Value: strconv.Itoa(u.Rating)},
		},
		Key: map[string]types.AttributeValue{
			"ItemName": &types.AttributeValueMemberS{Value: u.ItemName},
			"UserName": &types.AttributeValueMemberS{Value: u.UserName},
		},
		TableName:        aws.String(t.Name),
		UpdateExpression: aws.String("SET Rating = :rating"),
	}
	_, err := t.Client.UpdateItem(context.TODO(), input)
	return err
}

func (t UserScoreTable) GetUserScore(itemName, userName string) (UserScore, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"ItemName": &types.AttributeValueMemberS{Value: itemName},
			"UserName": &types.AttributeValueMemberS{Value: userName},
		},
		TableName: aws.String(t.Name),
	}
	output, err := t.Client.GetItem(context.TODO(), input)
	if err != nil {
		return UserScore{}, err
	}
	if output.Item == nil {
		return UserScore{}, fmt.Errorf("no user score found for item %s and user %s", itemName, userName)
	}
	rating, err := strconv.Atoi(output.Item["Rating"].(*types.AttributeValueMemberN).Value)
	return UserScore{
		ItemName: output.Item["ItemName"].(*types.AttributeValueMemberS).Value,
		UserName: output.Item["UserName"].(*types.AttributeValueMemberS).Value,
		Rating:   rating,
	}, err
}

func (t UserScoreTable) GetUserRatings(userName string) ([]UserScore, error) {
	input := &dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userName": &types.AttributeValueMemberS{Value: userName},
		},
		KeyConditionExpression: aws.String("UserName = :userName"),
		TableName:              aws.String(t.Name),
	}
	output, err := t.Client.Query(context.TODO(), input)
	if err != nil {
		return nil, err
	}
	var ratings []UserScore
	for _, item := range output.Items {
		rating, err := strconv.Atoi(item["Rating"].(*types.AttributeValueMemberN).Value)
		if err != nil {
			return nil, err
		}
		ratings = append(ratings, UserScore{
			ItemName: item["ItemName"].(*types.AttributeValueMemberS).Value,
			UserName: item["UserName"].(*types.AttributeValueMemberS).Value,
			Rating:   rating,
		})
	}
	return ratings, nil
}

type GlobalScoreTable Table

type GlobalScore struct {
	ItemName string `json:"item_name"`
	Score    int    `json:"score"`
}

func CreateGlobalScoreTable(client *dynamodb.Client) (Table, error) {
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("ItemName"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("ItemName"),
				KeyType:       types.KeyTypeHash,
			},
		},
		TableName:   aws.String("GlobalScores"),
		BillingMode: types.BillingModePayPerRequest,
	}
	_, err := client.CreateTable(context.TODO(), input)
	if err != nil {
		return Table{}, err
	}
	return Table{Name: "GlobalScores", Client: client}, nil
}

func (t GlobalScoreTable) PutGlobalScore(g GlobalScore) error {
	input := &dynamodb.PutItemInput{
		Item: map[string]types.AttributeValue{
			"ItemName": &types.AttributeValueMemberS{Value: g.ItemName},
			"Score":    &types.AttributeValueMemberN{Value: strconv.Itoa(g.Score)},
		},
		TableName: aws.String(t.Name),
	}
	_, err := t.Client.PutItem(context.TODO(), input)
	return err
}

func (t GlobalScoreTable) UpdateGlobalScore(g GlobalScore) error {
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":score": &types.AttributeValueMemberN{Value: strconv.Itoa(g.Score)},
		},
		Key: map[string]types.AttributeValue{
			"ItemName": &types.AttributeValueMemberS{Value: g.ItemName},
		},
		TableName:        aws.String(t.Name),
		UpdateExpression: aws.String("SET Score = :score"),
	}
	_, err := t.Client.UpdateItem(context.TODO(), input)
	return err
}

func (t GlobalScoreTable) GetGlobalScore(itemName string) (GlobalScore, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"ItemName": &types.AttributeValueMemberS{Value: itemName},
		},
		TableName: aws.String(t.Name),
	}
	output, err := t.Client.GetItem(context.TODO(), input)
	if err != nil {
		return GlobalScore{}, err
	}
	if output.Item == nil {
		return GlobalScore{}, fmt.Errorf("no global score found for item %s", itemName)
	}
	score, err := strconv.Atoi(output.Item["Score"].(*types.AttributeValueMemberN).Value)
	return GlobalScore{
		ItemName: output.Item["ItemName"].(*types.AttributeValueMemberS).Value,
		Score:    score,
	}, err
}
