package server

import (
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/quevivasbien/ranker-backend/database"
)

type PasswordMismatchError struct{}

func (e PasswordMismatchError) Error() string {
	return "incorrect password"
}

func GetToken(user database.User) (string, error) {
	secret := os.Getenv("RANKER_JWT_SECRET")
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user,
	})
	return claims.SignedString([]byte(secret))
}

func CheckToken(token string) (string, error) {
	secret := os.Getenv("RANKER_JWT_SECRET")
	t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}
	user := t.Claims.(jwt.MapClaims)["sub"].(string)
	return user, nil
}

// Login returns a JWT token for the user if the username and password are correct
func Login(db database.Database, username string, password string) (string, error) {
	user, err := db.Users.GetUser(username)
	if err != nil {
		return "", err
	}
	if user.Password != password {
		return "", PasswordMismatchError{}
	}
	return GetToken(user)
}
