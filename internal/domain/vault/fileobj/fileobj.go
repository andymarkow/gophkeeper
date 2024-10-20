// Package fileobj provides the domain model for file objects.
package fileobj

import (
	"encoding/hex"
	"fmt"
	"time"
)

// File represents a file object.
type File struct {
	id        string
	userID    string
	name      string
	location  string
	checksum  string
	salt      string
	iv        string
	size      int64
	metadata  map[string]string
	createdAt time.Time
	updatedAt time.Time
}

func CreateEmptyFileEntry(id, userID string, metadata map[string]string) (*File, error) {
	return NewFile(id, userID, "", "", "", "", "", 0, metadata, time.Now(), time.Now())
}

// NewFile creates a new file object.
func NewFile(id, userID, name, location, checksum, salt, iv string, size int64, metadata map[string]string,
	createdAt, updatedAt time.Time) (*File, error) {
	if id == "" {
		return nil, fmt.Errorf("id must not be empty")
	}

	if userID == "" {
		return nil, fmt.Errorf("user id must not be empty")
	}

	// if name == "" {
	// 	return nil, fmt.Errorf("name must not be empty")
	// }

	return &File{
		id:        id,
		userID:    userID,
		name:      name,
		location:  location,
		checksum:  checksum,
		salt:      salt,
		iv:        iv,
		size:      size,
		metadata:  metadata,
		createdAt: createdAt,
		updatedAt: updatedAt,
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

// SetName sets the name of the file object.
func (f *File) SetName(name string) {
	f.name = name
}

// Location returns the download url of the file object.
func (f *File) Location() string {
	return f.location
}

// Checksum returns the checksum of the file object.
func (f *File) Checksum() string {
	return f.checksum
}

// Salt returns the salt of the file object.
func (f *File) Salt() string {
	return f.salt
}

// SaltBytes returns the salt bytes of the file object.
func (f *File) SaltBytes() ([]byte, error) {
	saltBytes, err := hex.DecodeString(f.salt)
	if err != nil {
		return nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	return saltBytes, nil
}

// IV returns the initialization vector of the file object.
func (f *File) IV() string {
	return f.iv
}

// IVBytes returns the initialization vector bytes of the file object.
func (f *File) IVBytes() ([]byte, error) {
	ivBytes, err := hex.DecodeString(f.iv)
	if err != nil {
		return nil, fmt.Errorf("failed to decode iv: %w", err)
	}

	return ivBytes, nil
}

// Size returns the size of the file object.
func (f *File) Size() int64 {
	return f.size
}

// Metadata returns the metadata of the file object.
func (f *File) Metadata() map[string]string {
	return f.metadata
}

// AddMetadata adds metadata to the file object.
func (f *File) AddMetadata(metadata map[string]string) {
	for k, v := range metadata {
		f.metadata[k] = v
	}
}

// CreatedAt returns the create at of the file object.
func (f *File) CreatedAt() time.Time {
	return f.createdAt
}

// UpdatedAt returns the update at of the file object.
func (f *File) UpdatedAt() time.Time {
	return f.updatedAt
}

// SetUpdatedAt sets the update at of the file object.
func (f *File) SetUpdatedAt(t time.Time) {
	f.updatedAt = t
}
