// Package generic provides the domain model for generic secrets.
package generic

import (
	"fmt"
	"time"
)

// Generic represents a generic secret.
type Generic struct {
	id       string
	userID   string
	metadata map[string]string
	createAt time.Time
	data     *Data
}

// Data represents secret data.
type Data struct {
	content []byte
}

// NewGeneric creates a new generic secret.
func NewGeneric(id, userID string, metadata map[string]string, createAt time.Time, data *Data) (*Generic, error) {
	if id == "" {
		return nil, fmt.Errorf("id must not be empty")
	}

	if userID == "" {
		return nil, fmt.Errorf("user id must not be empty")
	}

	return &Generic{
		id:       id,
		userID:   userID,
		metadata: metadata,
		createAt: createAt,
		data:     data,
	}, nil
}

// ID returns the id of the secret.
func (s *Generic) ID() string {
	return s.id
}

// UserID returns the user login of the secret.
func (s *Generic) UserID() string {
	return s.userID
}

// Metadata returns the metadata of the secret.
func (s *Generic) Metadata() map[string]string {
	return s.metadata
}

// CreateAt returns the create at of the secret.
func (s *Generic) CreateAt() time.Time {
	return s.createAt
}

// Data returns the data of the secret.
func (s *Generic) Data() *Data {
	return s.data
}

// Content returns the key of the secret.
func (d *Data) Content() []byte {
	return d.content
}
