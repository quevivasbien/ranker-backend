package server

import (
	"fmt"
	"math"
	"math/rand"

	. "github.com/quevivasbien/ranker-backend/database"
)

const DEFAULT_ELO = 1000
const ELO_K = 64

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

func itemsWithFewestVotes(userScores []UserScore, excludeItem string) []string {
	items := []string{}
	minVotes := -1
	for _, userScore := range userScores {
		if userScore.ItemName == excludeItem {
			continue
		}
		if minVotes < 0 || userScore.NumVotes < minVotes {
			minVotes = userScore.NumVotes
			items = []string{}
		}
		if userScore.NumVotes == minVotes {
			items = append(items, userScore.ItemName)
		}
	}
	return items
}

func select1ItemForComparison(userScores []UserScore, item1 string) (string, string, error) {
	items := itemsWithFewestVotes(userScores, item1)
	i := rand.Intn(len(items))
	item2 := items[i]
	return item1, item2, nil
}

func select2ItemsForComparison(userScores []UserScore) (string, string, error) {
	items := itemsWithFewestVotes(userScores, "")
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
		return select1ItemForComparison(userScores, item1)
	}
}

// returns names of two items for user to compare with each other
func GetItemsForComparison(db Database, user string) (string, string, error) {
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
		return select1ItemForComparison(userScores, item1)
	}
}

func computeScoreChanges(score1 int, score2 int, winner1 bool) (int, int) {
	expected1 := 1 / (1 + math.Pow(10, float64(score2-score1)/400))
	expected2 := 1 / (1 + math.Pow(10, float64(score1-score2)/400))
	if winner1 {
		newScore1 := score1 + int(ELO_K*(1-expected1))
		newScore2 := score2 + int(ELO_K*(0-expected2))
		return newScore1, newScore2
	} else {
		newScore1 := score1 + int(ELO_K*(0-expected1))
		newScore2 := score2 + int(ELO_K*(1-expected2))
		return newScore1, newScore2
	}
}

func getOrCreateUserScore(db Database, item, user string) (UserScore, error) {
	userScore, err := db.UserScores.GetUserScore(item, user)
	if err == nil {
		return userScore, nil
	}
	if _, ok := err.(NotFoundError); ok {
		userScore = UserScore{ItemName: item, UserName: user, Rating: DEFAULT_ELO, NumVotes: 0}
		err = db.UserScores.PutUserScore(userScore)
		if err != nil {
			return userScore, fmt.Errorf("error creating user score in db: %v", err)
		}
		return userScore, nil
	}
	return userScore, fmt.Errorf("error getting user score from db: %v", err)
}

func getOrCreateGlobalScore(db Database, item string) (GlobalScore, error) {
	globalScore, err := db.GlobalScores.GetGlobalScore(item)
	if err == nil {
		return globalScore, nil
	}
	if _, ok := err.(NotFoundError); ok {
		globalScore = GlobalScore{ItemName: item, Rating: DEFAULT_ELO, NumVotes: 0}
		err = db.GlobalScores.PutGlobalScore(globalScore)
		if err != nil {
			return globalScore, fmt.Errorf("error creating global score in db: %v", err)
		}
		return globalScore, nil
	}
	return globalScore, fmt.Errorf("error getting global score from db: %v", err)
}

func ProcessUserChoice(db Database, user string, item1 string, item2 string, choice string) error {
	if choice != item1 && choice != item2 {
		return fmt.Errorf("invalid choice: %s", choice)
	}

	// compute user score updates
	userScore1, err := getOrCreateUserScore(db, item1, user)
	if err != nil {
		return err
	}
	userScore2, err := getOrCreateUserScore(db, item2, user)
	if err != nil {
		return err
	}
	userScore1.NumVotes++
	userScore2.NumVotes++
	winner1 := choice == item1
	userScore1.Rating, userScore2.Rating = computeScoreChanges(userScore1.Rating, userScore2.Rating, winner1)
	err = db.UserScores.UpdateUserScore(userScore1)
	if err != nil {
		return fmt.Errorf("error updating user score in db: %v", err)
	}
	db.UserScores.UpdateUserScore(userScore2)
	if err != nil {
		return fmt.Errorf("error updating user score in db: %v", err)
	}

	// compute global score updates
	globalScore1, err := getOrCreateGlobalScore(db, item1)
	if err != nil {
		return fmt.Errorf("error getting global score from db: %v", err)
	}
	globalScore2, err := getOrCreateGlobalScore(db, item2)
	if err != nil {
		return fmt.Errorf("error getting global score from db: %v", err)
	}
	globalScore1.NumVotes++
	globalScore2.NumVotes++
	globalScore1.Rating, globalScore2.Rating = computeScoreChanges(globalScore1.Rating, globalScore2.Rating, winner1)
	db.GlobalScores.UpdateGlobalScore(globalScore1)
	if err != nil {
		return fmt.Errorf("error updating global score in db: %v", err)
	}
	db.GlobalScores.UpdateGlobalScore(globalScore2)
	if err != nil {
		return fmt.Errorf("error updating global score in db: %v", err)
	}

	return nil
}
