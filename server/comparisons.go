package server

import (
	"fmt"
	"math/rand"

	. "github.com/quevivasbien/ranker-backend/database"
)

func containsItem(userScores []UserScore, itemName string) bool {
	for _, userScore := range userScores {
		if userScore.ItemName == itemName {
			return true
		}
	}
	return false
}

// returns names of unranked items
func getUnrankedItems(allItems []Item, userScores []UserScore) []string {
	var unrankedItems []string
	for _, item := range allItems {
		if !containsItem(userScores, item.Name) {
			unrankedItems = append(unrankedItems, item.Name)
		}
	}
	return unrankedItems
}

func itemsWithFewestVotes(userScores []UserScore, excludeIndex int) []string {
	items := []string{}
	minVotes := userScores[0].NumVotes
	for i, userScore := range userScores {
		if i == excludeIndex {
			continue
		}
		if userScore.NumVotes < minVotes {
			minVotes = userScore.NumVotes
			items = []string{}
		}
		if userScore.NumVotes == minVotes {
			items = append(items, userScore.ItemName)
		}
	}
	return items
}

func select1ItemForComparison(userScores []UserScore, item1 string, excludeIndex int) (string, string, error) {
	items := itemsWithFewestVotes(userScores, excludeIndex)
	i := rand.Intn(len(items))
	item2 := items[i]
	return item1, item2, nil
}

func select2ItemsForComparison(userScores []UserScore) (string, string, error) {
	items := itemsWithFewestVotes(userScores, -1)
	i := rand.Intn(len(items))
	item1 := items[i]
	if len(items) >= 2 {
		j := rand.Intn(len(items) - 1)
		if j >= i {
			j++
		}
		item2 := items[j]
		return item1, item2, nil
	} else {
		return select1ItemForComparison(userScores, item1, i)
	}
}

// returns names of two items for user to compare with each other
func getItemsForComparison(db Database, user string) (string, string, error) {
	allItems, err := db.Items.AllItems()
	if err != nil {
		return "", "", fmt.Errorf("error getting list of items from db: %v", err)
	}
	if len(allItems) < 2 {
		return "", "", fmt.Errorf("not enough items in db to compare")
	}
	userScores, err := db.UserScores.GetUserScores(user)
	if err != nil {
		return "", "", fmt.Errorf("error getting user scores from db: %v", err)
	}
	var item1, item2 string
	unrankedItems := getUnrankedItems(allItems, userScores)
	if len(unrankedItems) >= 1 {
		item1 = unrankedItems[0]
	} else {
		return select2ItemsForComparison(userScores)
	}
	if len(unrankedItems) >= 2 {
		item2 = unrankedItems[1]
		return item1, item2, nil
	} else {
		return select1ItemForComparison(userScores, item1, -1)
	}
}
