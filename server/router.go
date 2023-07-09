package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/quevivasbien/ranker-backend/database"
)

// send the right HTTP status code for an error
func setHTTPError(w http.ResponseWriter, err error) {
	var statusCode int
	if _, ok := err.(database.NotFoundError); ok {
		statusCode = http.StatusNotFound
	} else if _, ok := err.(PasswordMismatchError); ok {
		statusCode = http.StatusUnauthorized
	} else if _, ok := err.(TokenMissingError); ok {
		statusCode = http.StatusUnauthorized
	} else if _, ok := err.(InsufficientPermissionsError); ok {
		statusCode = http.StatusForbidden
	} else {
		statusCode = http.StatusInternalServerError
	}
	log.Printf("Error: %s; wrote status code %d", err.Error(), statusCode)
	w.WriteHeader(statusCode)
	w.Write([]byte(err.Error()))
}

// create handler for /items endpoint
func handleItems(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get list of all items
		if r.Method == "GET" {
			items, err := db.Items.AllItems()
			if err != nil {
				setHTTPError(w, err)
				return
			}

			bytes, err := json.Marshal(items)

			w.WriteHeader(http.StatusOK)
			w.Write(bytes)
			return
		}

		// create a new item
		if r.Method == "POST" {
			// require jwt token
			_, err := VerifyUser(r)
			if err != nil {
				setHTTPError(w, err)
				return
			}

			var item database.Item
			err = json.NewDecoder(r.Body).Decode(&item)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}

			err = db.Items.PutItem(item)
			if err != nil {
				setHTTPError(w, err)
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		}
	}
}

// create handler for /items/{name} endpoint
func handleItem(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["item"]

		// get a single item
		if r.Method == "GET" {
			item, err := db.Items.GetItem(name)
			if err != nil {
				setHTTPError(w, err)
				return
			}

			bytes, err := json.Marshal(item)

			w.WriteHeader(http.StatusOK)
			w.Write(bytes)
			return
		}

		// delete an item
		if r.Method == "DELETE" {
			// require jwt token and admin status
			_, err := VerifyAdmin(r)
			if err != nil {
				setHTTPError(w, err)
				return
			}

			err = db.Items.DeleteItem(name)
			if err != nil {
				setHTTPError(w, err)
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		}
	}
}

// create handler for /users endpoint
func handleUsers(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// get list of all users
		if r.Method == "GET" {

			// require jwt token and admin status
			_, err := VerifyAdmin(r)
			if err != nil {
				setHTTPError(w, err)
				return
			}

			users, err := db.Users.AllUsers()
			if err != nil {
				setHTTPError(w, err)
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
				setHTTPError(w, err)
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		}
	}
}

// create handler for /users/{name} endpoint
func handleUser(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]

		// require jwt token and admin status or matching username
		username, err := VerifyUser(r)
		if err != nil {
			setHTTPError(w, err)
			return
		}
		if username != name && username != "admin" {
			setHTTPError(w, InsufficientPermissionsError{})
			return
		}

		// get a single user
		if r.Method == "GET" {
			user, err := db.Users.GetUser(name)
			if err != nil {
				setHTTPError(w, err)
				return
			}

			bytes, err := json.Marshal(user)

			w.WriteHeader(http.StatusOK)
			w.Write(bytes)
			return
		}

		// delete a user
		if r.Method == "DELETE" {
			err = db.Users.DeleteUser(name)
			if err != nil {
				setHTTPError(w, err)
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		}
	}
}

type comparisonResponse struct {
	Item1  string `json:"item1"`
	Item2  string `json:"item2"`
	Winner string `json:"winner"`
}

// create handler for /compare endpoint
func handleCompare(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// require jwt token
		username, err := VerifyUser(r)
		if err != nil {
			setHTTPError(w, err)
			return
		}

		// get items for comparison
		if r.Method == "GET" {
			item1, item2, err := GetItemsForComparison(db, username)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			bytes, err := json.Marshal([]string{item1, item2})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(bytes)
			return
		}

		// send the result of a comparison
		if r.Method == "POST" {
			var response comparisonResponse
			err := json.NewDecoder(r.Body).Decode(&response)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
			err = ProcessUserChoice(db, username, response.Item1, response.Item2, response.Winner)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}
			return
		}
	}
}

// create handler for /scores/{item} endpoint
func handleGlobalScore(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		itemName := vars["item"]

		// get the score for a single item
		if r.Method == "GET" {
			globalScore, err := db.GlobalScores.GetGlobalScore(itemName)
			if err != nil {
				setHTTPError(w, err)
				return
			}
			bytes, err := json.Marshal(globalScore)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(bytes)
			return
		}
	}
}

// create handler for /scores/{item}/{user} endpoint
func handleUserScore(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		itemName := vars["item"]
		name := vars["user"]

		// require jwt token and admin status or matching username
		username, err := VerifyUser(r)
		if err != nil {
			setHTTPError(w, err)
			return
		}
		if username != name && username != "admin" {
			setHTTPError(w, InsufficientPermissionsError{})
			return
		}

		// get the score for a single item
		if r.Method == "GET" {
			userScore, err := db.UserScores.GetUserScore(itemName, name)
			if err != nil {
				setHTTPError(w, err)
				return
			}
			bytes, err := json.Marshal(userScore)
			if err != nil {
				setHTTPError(w, err)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(bytes)
			return
		}
	}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// create handler for /login endpoint
func handleLogin(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var request loginRequest
			err := json.NewDecoder(r.Body).Decode(&request)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
			token, err := Login(db, request.Username, request.Password)
			if err != nil {
				setHTTPError(w, err)
				return
			}
			bytes, err := json.Marshal(token)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(bytes)
		}
	}

}

func CreateRouter() (http.Handler, error) {
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
	r.HandleFunc("/items/{item}", handleItem(db)).Methods("GET", "DELETE")

	r.HandleFunc("/users", handleUsers(db)).Methods("GET", "POST")
	r.HandleFunc("/users/{name}", handleUser(db)).Methods("GET", "DELETE")

	r.HandleFunc("/compare", handleCompare(db)).Methods("GET", "POST")

	r.HandleFunc("/scores/{item}", handleGlobalScore(db)).Methods("GET")
	r.HandleFunc("/scores/{item}/{user}", handleUserScore(db)).Methods("GET")

	r.HandleFunc("/login", handleLogin(db)).Methods("POST")

	handler := cors.New(
		cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowCredentials: true,
			AllowedHeaders:   []string{"Authorization", "Content-Type"},
		},
	).Handler(r)

	return handler, nil
}
