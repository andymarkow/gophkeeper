package credrepo

import "fmt"

var (
	ErrSecretAlreadyExists = fmt.Errorf("credential secret already exists")
	ErrSecretNotFound      = fmt.Errorf("credential secret not found")
)
