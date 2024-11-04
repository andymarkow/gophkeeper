package userapi

import "fmt"

var (
	ErrUserAlreadyExists      = fmt.Errorf("user already exists")
	ErrUserNotFound           = fmt.Errorf("user not found")
	ErrUserInvalidCredentials = fmt.Errorf("user invalid credentials")
)
