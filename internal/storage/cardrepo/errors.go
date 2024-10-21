package cardrepo

import "fmt"

var (
	ErrSecretAlreadyExists = fmt.Errorf("bank card secret entry already exists")
	ErrSecretNotFound      = fmt.Errorf("bank card secret entry not found")
)
