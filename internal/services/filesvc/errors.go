package filesvc

import "errors"

var (
	ErrFileEntryNotFound  = errors.New("file not found")
	ErrFileObjectNotFound = errors.New("file object not found")
)
