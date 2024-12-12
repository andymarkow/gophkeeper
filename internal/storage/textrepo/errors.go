package textrepo

import "fmt"

var (
	ErrSecretAlreadyExists = fmt.Errorf("text secret entry already exists")
	ErrSecretNotFound      = fmt.Errorf("text secret entry not found")
)
