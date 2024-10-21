package filesvc

import "errors"

var (
	ErrFileEntryAlreadyExists = errors.New("file entry already exists")
	ErrFileEntryNotFound      = errors.New("file entry not found")
	ErrFileObjectNotFound     = errors.New("file object not found")
)
