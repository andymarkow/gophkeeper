// Package user provides the domain model for a user.
package user

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user.
type User struct {
	login    string
	password string
}

// NewUser creates a new user.
func NewUser(login, password string) (*User, error) {
	if login == "" {
		return nil, fmt.Errorf("login must not be empty")
	}

	if password == "" {
		return nil, fmt.Errorf("password must not be empty")
	}

	return &User{
		login:    login,
		password: password,
	}, nil
}

// CreateUser creates a new user and generates the password hash.
func CreateUser(login, password string) (*User, error) {
	if password == "" {
		return nil, fmt.Errorf("password must not be empty")
	}

	pwdHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("bcrypt.GenerateFromPassword: %w", err)
	}

	return NewUser(login, string(pwdHash))
}

// Login returns the login of the user.
func (u *User) Login() string {
	return u.login
}

// Password returns the password of the user.
func (u *User) Password() string {
	return u.password
}
