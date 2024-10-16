package filerepo

import "fmt"

var (
	ErrFileAlreadyExists = fmt.Errorf("file already exists")
	ErrFileNotFound      = fmt.Errorf("file not found")
)
