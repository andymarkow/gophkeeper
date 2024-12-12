package filerepo

import "fmt"

var (
	ErrSecretAlreadyExists = fmt.Errorf("secret entry already exists")
	ErrSecretNotFound      = fmt.Errorf("secret entry not found")
)
