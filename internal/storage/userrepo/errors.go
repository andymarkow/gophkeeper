package userrepo

import "errors"

var (
	ErrUsrAlreadyExists = errors.New("user already exists")
	ErrUsrNotFound      = errors.New("user not found")
)
