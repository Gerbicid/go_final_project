package users

import (
	"fmt"
	"os"
	"sync"
)

var (
	password string
	once     sync.Once
)

type User struct {
	PasswordString string `json:"password,omitempty"`
}

func initPassword() {
	var ok bool
	password, ok = os.LookupEnv("TODO_PASSWORD")
	if !ok {
		panic("TODO_PASSWORD environment variable not set")
	}
}

func (u User) GetPasswordFromEnv() (string, error) {
	once.Do(initPassword)
	if password == "" {
		return "", fmt.Errorf("the password value is empty in the .env file")
	}
	return password, nil
}

func (u User) GetPassword() string {
	return u.PasswordString
}
