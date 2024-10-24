package filesvc

import "errors"

var (
	ErrSecretEntryAlreadyExists = errors.New("file secret entry already exists")
	ErrSecretEntryNotFound      = errors.New("file secret entry not found")
	ErrSecretObjectNotFound     = errors.New("file secret object not found")
)
