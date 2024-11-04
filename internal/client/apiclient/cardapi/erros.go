package cardapi

import "errors"

var (
	ErrSecretNotFound      = errors.New("secret entry not found")
	ErrSecretAlreadyExists = errors.New("secret entry already exists")
)
