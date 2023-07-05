package main

import (
	"net/http"

	"github.com/quevivasbien/ranker-backend/server"
)

func main() {
	router, err := server.CreateRouter()
	if err != nil {
		panic(err)
	}
	http.ListenAndServe(":8080", router)
}
