package textrepo

import "fmt"

var (
	ErrSecretEntryAlreadyExists = fmt.Errorf("secret entry already exists")
	ErrSecretEntryNotFound      = fmt.Errorf("secret entry not found")
)
