// Package fileobj provides the domain model for file objects.
package fileobj

import (
	"fmt"
	"time"

	"net/url"
)

// File represents a file object.
type File struct {
	id        string
	userID    string
	name      string
	checksum  string
	metadata  map[string]string
	createdAt time.Time
	updatedAt time.Time
	location  *url.URL
}

// NewFile creates a new file object.
func NewFile(id, userID, name string, metadata map[string]string,
	createdAt, updatedAt time.Time, location *url.URL, checksum string) (*File, error) {
	if id == "" {
		return nil, fmt.Errorf("id must not be empty")
	}

	if userID == "" {
		return nil, fmt.Errorf("user id must not be empty")
	}

	if name == "" {
		return nil, fmt.Errorf("name must not be empty")
	}

	return &File{
		id:        id,
		userID:    userID,
		name:      name,
		metadata:  metadata,
		createdAt: createdAt,
		updatedAt: updatedAt,
		location:  location,
		checksum:  checksum,
	}, nil
}

// ID returns the id of the file object.
func (f *File) ID() string {
	return f.id
}

// UserID returns the user login of the file object.
func (f *File) UserID() string {
	return f.userID
}

// Name returns the name of the file object.
func (f *File) Name() string {
	return f.name
}

// Metadata returns the metadata of the file object.
func (f *File) Metadata() map[string]string {
	return f.metadata
}

// CreatedAt returns the create at of the file object.
func (f *File) CreatedAt() time.Time {
	return f.createdAt
}

// UpdatedAt returns the update at of the file object.
func (f *File) UpdatedAt() time.Time {
	return f.updatedAt
}

// Location returns the download url of the file object.
func (f *File) Location() *url.URL {
	return f.location
}

// Checksum returns the checksum of the file object.
func (f *File) Checksum() string {
	return f.checksum
}
