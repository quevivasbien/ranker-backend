package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/quevivasbien/ranker-backend/database"
)

// create handler for /items endpoint
func handleItems(db database.Database) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

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
		if r.Method == "POST" {
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

// create handler for /items/{name} endpoint
func handleItem(db database.Database) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// get a single item
		if r.Method == "GET" {
			item, err := db.Items.GetItem(name)
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

		// delete an item
		if r.Method == "DELETE" {
			err := db.Items.DeleteItem(name)
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

// create handler for /users endpoint
func handleUsers(db database.Database) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// get list of all users
		if r.Method == "GET" {
			users, err := db.Users.AllUsers()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			bytes, err := json.Marshal(users)

			w.WriteHeader(http.StatusOK)
			w.Write(bytes)
			return
		}

		// add a new user
		if r.Method == "POST" {
			var user database.User
			err := json.NewDecoder(r.Body).Decode(&user)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}

			err = db.Users.PutUser(user)
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

// create handler for /users/{name} endpoint
func handleUser(db database.Database) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		vars := mux.Vars(r)
		name := vars["name"]

		// get a single user
		if r.Method == "GET" {
			user, err := db.Users.GetUser(name)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			bytes, err := json.Marshal(user)

			w.WriteHeader(http.StatusOK)
			w.Write(bytes)
			return
		}

		// delete a user
		if r.Method == "DELETE" {
			err := db.Users.DeleteUser(name)
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

	r.HandleFunc("/items", handleItems(db)).Methods("GET", "POST")
	r.HandleFunc("/items/{name}", handleItem(db)).Methods("GET", "DELETE")

	r.HandleFunc("/users", handleUsers(db)).Methods("GET", "POST")
	r.HandleFunc("/users/{name}", handleUser(db)).Methods("GET", "DELETE")

	return r, nil
}
