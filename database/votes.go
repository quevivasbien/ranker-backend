package database

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// a vote on an item
type UserScore struct {
	ItemID string `json:"item_id"`
	UserID string `json:"user_id"`
}

func createUserScoreTable(client *dynamodb.Client) (Table, error) {
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
		TableName: aws.String("UserScores"),
	}
	_, err := client.CreateTable(context.TODO(), input)
	if err != nil {
		return Table{}, err
	}
	return Table{Name: "UserScores", Client: client}, nil
}

func (t Table) PutUserScore(u UserScore) error {
	input := &dynamodb.PutItemInput{
		Item: map[string]types.AttributeValue{
			"ItemID": &types.AttributeValueMemberS{Value: u.ItemID},
			"UserID": &types.AttributeValueMemberS{Value: u.UserID},
		},
		TableName: aws.String(t.Name),
	}
	_, err := t.Client.PutItem(context.TODO(), input)
	return err
}

func (t Table) GetVote(itemID, userID string) (UserScore, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"ItemID": &types.AttributeValueMemberS{Value: itemID},
			"UserID": &types.AttributeValueMemberS{Value: userID},
		},
		TableName: aws.String(t.Name),
	}
	output, err := t.Client.GetItem(context.Background(), input)
	if err != nil {
		return UserScore{}, err
	}
	return UserScore{
		ItemID: output.Item["ItemID"].(*types.AttributeValueMemberS).Value,
		UserID: output.Item["UserID"].(*types.AttributeValueMemberS).Value,
	}, nil
}
