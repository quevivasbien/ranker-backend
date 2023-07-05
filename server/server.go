package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/quevivasbien/ranker-backend/database"
)

func handleItems(db database.Database) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get list of all items
		if r.Method == "GET" {
			items, err := db.Items.AllItems()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			bytes, err := json.Marshal(items)

			w.WriteHeader(http.StatusOK)
			w.Write(bytes)
			return
		}

		// create a new item
		if r.Method == "PUT" {
			var item database.Item
			err := json.NewDecoder(r.Body).Decode(&item)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}

			err = db.Items.PutItem(item)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleItem(db database.Database) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		if r.Method == "GET" {
			item, err := db.Items.GetItem(id)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			bytes, err := json.Marshal(item)

			w.WriteHeader(http.StatusOK)
			w.Write(bytes)
			return
		}

		if r.Method == "DELETE" {
			err := db.Items.DeleteItem(id)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		}
	}
}

func CreateRouter() (*mux.Router, error) {
	r := mux.NewRouter()
	client, err := database.GetClient("us-east-1")
	if err != nil {
		return nil, err
	}
	db, err := database.GetDatabase(client)
	if err != nil {
		return nil, err
	}

	r.HandleFunc("/items", handleItems(db)).Methods("GET", "PUT")
	r.HandleFunc("/items/{id}", handleItem(db)).Methods("GET", "DELETE")

	return r, nil
}
