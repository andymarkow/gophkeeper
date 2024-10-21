// Package user provides the domain model for a user.
package user

import (
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user.
type User struct {
	id       string
	login    string
	password string
}

// NewUser creates a new user.
func NewUser(id, login, password string) (*User, error) {
	if id == "" {
		return nil, fmt.Errorf("id must not be empty")
	}

	if login == "" {
		return nil, fmt.Errorf("login must not be empty")
	}

	if password == "" {
		return nil, fmt.Errorf("password must not be empty")
	}

	return &User{
		id:       id,
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

	return NewUser(uuid.New().String(), login, string(pwdHash))
}

// ID returns the ID of the user.
func (u *User) ID() string {
	return u.id
}

// Login returns the login of the user.
func (u *User) Login() string {
	return u.login
}

// Password returns the password of the user.
func (u *User) Password() string {
	return u.password
}
