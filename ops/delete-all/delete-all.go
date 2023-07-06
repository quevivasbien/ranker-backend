package main

import "github.com/quevivasbien/ranker-backend/database"

func main() {
	client, err := database.GetClient("us-east-1")
	if err != nil {
		panic(err)
	}
	tableNames, err := database.ListTables(client)
	if err != nil {
		panic(err)
	}
	for _, tableName := range tableNames {
		err := database.DeleteTable(client, tableName)
		if err != nil {
			panic(err)
		}
	}
}
