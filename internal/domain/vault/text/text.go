// Package text provides the domain model for the text data.
package text

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Secret represents a secret.
type Secret struct {
	id        string
	name      string
	userID    string
	metadata  map[string]string
	createdAt time.Time
	updatedAt time.Time
	info      *ContentInfo
}

// ContentInfo represents a secret content info.
type ContentInfo struct {
	salt     string
	iv       string
	location string
	checksum string
}

// NewContentInfo creates a new secret content info.
func NewContentInfo(salt, iv, location, checksum string) *ContentInfo {
	return &ContentInfo{
		salt:     salt,
		iv:       iv,
		location: location,
		checksum: checksum,
	}
}

// NewSecret creates a new secret.
func NewSecret(id, name, userID string, metadata map[string]string,
	createdAt, updatedAt time.Time, info *ContentInfo) (*Secret, error) {
	if id == "" {
		return nil, fmt.Errorf("secret ID must not be empty")
	}

	if name == "" {
		return nil, fmt.Errorf("secret name must not be empty")
	}

	if userID == "" {
		return nil, fmt.Errorf("user ID must not be empty")
	}

	if metadata == nil {
		metadata = make(map[string]string)
	}

	if info == nil {
		info = &ContentInfo{}
	}

	return &Secret{
		id:        id,
		name:      name,
		userID:    userID,
		metadata:  metadata,
		createdAt: createdAt,
		updatedAt: updatedAt,
		info:      info,
	}, nil
}

// CreateSecret creates a new secret object.
func CreateSecret(name, userID string, metadata map[string]string) (*Secret, error) {
	return NewSecret(uuid.New().String(), name, userID, metadata, time.Now(), time.Now(), nil)
}

// ID returns the id of the secret.
func (s *Secret) ID() string {
	return s.id
}

// Name returns the name of the secret.
func (s *Secret) Name() string {
	return s.name
}

// UserID returns the user ID of the secret.
func (s *Secret) UserID() string {
	return s.userID
}

// Metadata returns the metadata of the secret.
func (s *Secret) Metadata() map[string]string {
	return s.metadata
}

// AddMetadata adds metadata to the secret.
func (s *Secret) AddMetadata(metadata map[string]string) {
	for k, v := range metadata {
		s.metadata[k] = v
	}
}

// CreatedAt returns the creation time of the secret.
func (s *Secret) CreatedAt() time.Time {
	return s.createdAt
}

// UpdatedAt returns the update time of the secret.
func (s *Secret) UpdatedAt() time.Time {
	return s.updatedAt
}

// SetUpdatedAt sets the update time of the secret.
func (s *Secret) SetUpdatedAt(tm time.Time) {
	s.updatedAt = tm
}

// ContentInfo returns the content info of the secret.
func (s *Secret) ContentInfo() *ContentInfo {
	return s.info
}

// SetContentInfo sets the content info of the secret.
func (s *Secret) SetContentInfo(info *ContentInfo) {
	s.info = info
}

// Salt returns the salt of the secret.
func (i *ContentInfo) Salt() string {
	return i.salt
}

// IV returns the initialisation vector of the secret.
func (i *ContentInfo) IV() string {
	return i.iv
}

// Location returns the location of the secret.
func (i *ContentInfo) Location() string {
	return i.location
}

// Checksum returns the checksum of the secret.
func (i *ContentInfo) Checksum() string {
	return i.checksum
}
