package textsvc

import "errors"

var (
	ErrSecretEntryAlreadyExists = errors.New("secret entry already exists")
	ErrSecretEntryNotFound      = errors.New("secret entry not found")
	ErrSecretObjectNotFound     = errors.New("secret object not found")
)
