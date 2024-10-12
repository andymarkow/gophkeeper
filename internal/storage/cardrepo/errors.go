package cardrepo

import "fmt"

var (
	ErrCardAlreadyExists = fmt.Errorf("bank card already exists")
	ErrCardNotFound      = fmt.Errorf("bank card not found")
)
