package credrepo

import "fmt"

var (
	ErrCredAlreadyExists = fmt.Errorf("credential already exists")
	ErrCredNotFound      = fmt.Errorf("credential not found")
)
