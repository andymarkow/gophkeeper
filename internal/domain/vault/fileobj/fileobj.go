// Package fileobj provides the domain model for file objects.
package fileobj

import (
	"fmt"
	"time"

	"net/url"
)

// File represents a file object.
type File struct {
	id       string
	userID   string
	name     string
	metadata map[string]string
	createAt time.Time
	updateAt time.Time
	url      url.URL
	checksum uint32
}

// NewFile creates a new file object.
func NewFile(id, userID, name string, metadata map[string]string, createAt, updateAt time.Time, url url.URL, checksum uint32) (*File, error) {
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
		id:       id,
		userID:   userID,
		name:     name,
		metadata: metadata,
		createAt: createAt,
		updateAt: updateAt,
		url:      url,
		checksum: checksum,
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

// CreateAt returns the create at of the file object.
func (f *File) CreateAt() time.Time {
	return f.createAt
}

// UpdateAt returns the update at of the file object.
func (f *File) UpdateAt() time.Time {
	return f.updateAt
}

// URL returns the download url of the file object.
func (f *File) URL() url.URL {
	return f.url
}

// Checksum returns the checksum of the file object.
func (f *File) Checksum() uint32 {
	return f.checksum
}
