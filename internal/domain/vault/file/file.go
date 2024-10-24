// Package file provides the domain model for file objects.
package file

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Secret represents a file secret.
type Secret struct {
	id        string
	name      string
	userID    string
	metadata  map[string]string
	createdAt time.Time
	updatedAt time.Time
	info      *ContentInfo
}

// ContentInfo represents the content info of a file.
type ContentInfo struct {
	fileName string
	location string
	checksum string
	salt     string
	iv       string
	size     int64
}

// NewContentInfo creates a new file info object.
func NewContentInfo(fileName, location, checksum, salt, iv string, size int64) (*ContentInfo, error) {
	return &ContentInfo{
		fileName: fileName,
		location: location,
		checksum: checksum,
		salt:     salt,
		iv:       iv,
		size:     size,
	}, nil
}

// NewSecret creates a new file object.
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

// CreateSecret creates a new file secret object.
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

// SetName sets the name of the secret.
func (s *Secret) SetName(name string) {
	s.name = name
}

// UserID returns the user id of the secret.
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

// CreatedAt returns the created at time of the secret.
func (s *Secret) CreatedAt() time.Time {
	return s.createdAt
}

// UpdatedAt returns the updated at time of the secret.
func (s *Secret) UpdatedAt() time.Time {
	return s.updatedAt
}

// SetUpdatedAt sets the updated at time of the secret.
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

// FileName returns the file name of the secret.
func (i *ContentInfo) FileName() string {
	return i.fileName
}

// SetFileName sets the file name of the secret.
func (i *ContentInfo) SetFileName(fileName string) {
	i.fileName = fileName
}

// Location returns the location of the secret.
func (i *ContentInfo) Location() string {
	return i.location
}

// SetLocation sets the location of the secret.
func (i *ContentInfo) SetLocation(location string) {
	i.location = location
}

// Checksum returns the checksum of the secret.
func (i *ContentInfo) Checksum() string {
	return i.checksum
}

// SetChecksum sets the checksum of the secret.
func (i *ContentInfo) SetChecksum(checksum string) {
	i.checksum = checksum
}

// Salt returns the salt of the secret.
func (i *ContentInfo) Salt() string {
	return i.salt
}

// SetSalt sets the salt of the secret.
func (i *ContentInfo) SetSalt(salt string) {
	i.salt = salt
}

// IV returns the iv of the secret.
func (i *ContentInfo) IV() string {
	return i.iv
}

// SetIV sets the iv of the secret.
func (i *ContentInfo) SetIV(iv string) {
	i.iv = iv
}

// Size returns the size of the secret.
func (i *ContentInfo) Size() int64 {
	return i.size
}

// SetSize sets the size of the secret.
func (i *ContentInfo) SetSize(size int64) {
	i.size = size
}
