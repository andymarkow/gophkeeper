package bankcard

import "fmt"

var (
	ErrItemAlreadyExists = fmt.Errorf("item already exists")
	ErrItemNotFound      = fmt.Errorf("item not found")
)
