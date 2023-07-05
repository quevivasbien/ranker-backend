package main

import (
	"go-aws/server"
	"net/http"
)

func main() {
	router, err := server.CreateRouter()
	if err != nil {
		panic(err)
	}
	http.ListenAndServe(":8080", router)
}
