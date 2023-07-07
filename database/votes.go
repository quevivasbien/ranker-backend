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
	ItemName string `json:"itemName"`
	UserName string `json:"userName"`
	Rating   int    `json:"rating"`
	NumVotes int    `json:"numVotes"`
}

func CreateUserScoreTable(client *dynamodb.Client) (UserScoreTable, error) {
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("UserName"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("ItemName"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("UserName"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("ItemName"),
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
			"NumVotes": &types.AttributeValueMemberN{Value: strconv.Itoa(u.NumVotes)},
		},
		TableName: aws.String(t.Name),
	}
	_, err := t.Client.PutItem(context.TODO(), input)
	return err
}

func (t UserScoreTable) UpdateUserScore(u UserScore) error {
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":rating":   &types.AttributeValueMemberN{Value: strconv.Itoa(u.Rating)},
			":numVotes": &types.AttributeValueMemberN{Value: strconv.Itoa(u.NumVotes)},
		},
		Key: map[string]types.AttributeValue{
			"ItemName": &types.AttributeValueMemberS{Value: u.ItemName},
			"UserName": &types.AttributeValueMemberS{Value: u.UserName},
		},
		TableName:        aws.String(t.Name),
		UpdateExpression: aws.String("SET Rating = :rating, NumVotes = :numVotes"),
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
		return UserScore{}, MakeNotFoundError(fmt.Sprintf("no user score found for item %s and user %s", itemName, userName))
	}
	rating, err := strconv.Atoi(output.Item["Rating"].(*types.AttributeValueMemberN).Value)
	if err != nil {
		return UserScore{}, err
	}
	numVotes, err := strconv.Atoi(output.Item["NumVotes"].(*types.AttributeValueMemberN).Value)
	if err != nil {
		return UserScore{}, err
	}
	return UserScore{
		ItemName: output.Item["ItemName"].(*types.AttributeValueMemberS).Value,
		UserName: output.Item["UserName"].(*types.AttributeValueMemberS).Value,
		Rating:   rating,
		NumVotes: numVotes,
	}, nil
}

func (t UserScoreTable) GetUserScores(userName string) ([]UserScore, error) {
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
		numVotes, err := strconv.Atoi(item["NumVotes"].(*types.AttributeValueMemberN).Value)
		if err != nil {
			return nil, err
		}
		ratings = append(ratings, UserScore{
			ItemName: item["ItemName"].(*types.AttributeValueMemberS).Value,
			UserName: item["UserName"].(*types.AttributeValueMemberS).Value,
			Rating:   rating,
			NumVotes: numVotes,
		})
	}
	return ratings, nil
}

type GlobalScoreTable Table

type GlobalScore struct {
	ItemName string `json:"itemName"`
	Rating   int    `json:"rating"`
	NumVotes int    `json:"numVotes"`
}

func CreateGlobalScoreTable(client *dynamodb.Client) (GlobalScoreTable, error) {
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
		return GlobalScoreTable{}, err
	}
	return GlobalScoreTable{Name: "GlobalScores", Client: client}, nil
}

func (t GlobalScoreTable) PutGlobalScore(g GlobalScore) error {
	input := &dynamodb.PutItemInput{
		Item: map[string]types.AttributeValue{
			"ItemName": &types.AttributeValueMemberS{Value: g.ItemName},
			"Rating":   &types.AttributeValueMemberN{Value: strconv.Itoa(g.Rating)},
			"NumVotes": &types.AttributeValueMemberN{Value: strconv.Itoa(g.NumVotes)},
		},
		TableName: aws.String(t.Name),
	}
	_, err := t.Client.PutItem(context.TODO(), input)
	return err
}

func (t GlobalScoreTable) UpdateGlobalScore(g GlobalScore) error {
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":rating":   &types.AttributeValueMemberN{Value: strconv.Itoa(g.Rating)},
			":numVotes": &types.AttributeValueMemberN{Value: strconv.Itoa(g.NumVotes)},
		},
		Key: map[string]types.AttributeValue{
			"ItemName": &types.AttributeValueMemberS{Value: g.ItemName},
		},
		TableName:        aws.String(t.Name),
		UpdateExpression: aws.String("set Rating = :rating, NumVotes = :numVotes"),
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
		return GlobalScore{}, MakeNotFoundError(fmt.Sprintf("no global score found for item %s", itemName))
	}
	rating, err := strconv.Atoi(output.Item["Rating"].(*types.AttributeValueMemberN).Value)
	if err != nil {
		return GlobalScore{}, err
	}
	numVotes, err := strconv.Atoi(output.Item["NumVotes"].(*types.AttributeValueMemberN).Value)
	if err != nil {
		return GlobalScore{}, err
	}
	return GlobalScore{
		ItemName: output.Item["ItemName"].(*types.AttributeValueMemberS).Value,
		Rating:   rating,
		NumVotes: numVotes,
	}, nil
}
