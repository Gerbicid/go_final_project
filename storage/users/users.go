package users

import (
	"fmt"
	"os"
)

type User struct {
	PasswordString string `json:"password,omitempty"`
}

func (u User) GetPasswordFromEnv() (string, error) {
	passwordEnv := os.Getenv("TODO_PASSWORD")
	if passwordEnv == "" {
		return "", fmt.Errorf("the password value is empty in the .env file is empty.")
	}
	return passwordEnv, nil
}

func (u User) GetPassword() string {
	return u.PasswordString
}
