package database

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type UserScoreTable Table

// a vote on an item
type UserScore struct {
	ItemID string `json:"item_id"`
	UserID string `json:"user_id"`
	Rating int    `json:"rating"`
}

func CreateUserScoreTable(client *dynamodb.Client) (UserScoreTable, error) {
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("ItemID"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("UserID"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("ItemID"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("UserID"),
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
			"ItemID": &types.AttributeValueMemberS{Value: u.ItemID},
			"UserID": &types.AttributeValueMemberS{Value: u.UserID},
			"Rating": &types.AttributeValueMemberN{Value: strconv.Itoa(u.Rating)},
		},
		TableName: aws.String(t.Name),
	}
	_, err := t.Client.PutItem(context.TODO(), input)
	return err
}

func (t UserScoreTable) GetUserScore(itemID, userID string) (UserScore, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"ItemID": &types.AttributeValueMemberS{Value: itemID},
			"UserID": &types.AttributeValueMemberS{Value: userID},
		},
		TableName: aws.String(t.Name),
	}
	output, err := t.Client.GetItem(context.TODO(), input)
	if err != nil {
		return UserScore{}, err
	}
	rating, err := strconv.Atoi(output.Item["Rating"].(*types.AttributeValueMemberN).Value)
	return UserScore{
		ItemID: output.Item["ItemID"].(*types.AttributeValueMemberS).Value,
		UserID: output.Item["UserID"].(*types.AttributeValueMemberS).Value,
		Rating: rating,
	}, err
}

func (t UserScoreTable) GetUserRatings(userID string) ([]UserScore, error) {
	input := &dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userID": &types.AttributeValueMemberS{Value: userID},
		},
		KeyConditionExpression: aws.String("UserID = :userID"),
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
			ItemID: item["ItemID"].(*types.AttributeValueMemberS).Value,
			UserID: item["UserID"].(*types.AttributeValueMemberS).Value,
			Rating: rating,
		})
	}
	return ratings, nil
}
