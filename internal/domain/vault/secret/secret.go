// Package secret provides the domain model for generic secrets.
package secret

import "fmt"

// Secret represents a generic secret.
type Secret struct {
	id        string
	userLogin string
	metadata  map[string]string
	data      *Data
}

// Data represents secret data.
type Data struct {
	key   string
	value string
}

// NewSecret creates a new generic secret.
func NewSecret(id, userLogin string, metadata map[string]string, data *Data) (*Secret, error) {
	if id == "" {
		return nil, fmt.Errorf("id must not be empty")
	}

	if userLogin == "" {
		return nil, fmt.Errorf("user login must not be empty")
	}

	if data.key == "" {
		return nil, fmt.Errorf("key must not be empty")
	}

	return &Secret{
		id:        id,
		userLogin: userLogin,
		metadata:  metadata,
		data:      data,
	}, nil
}

// ID returns the id of the secret.
func (s *Secret) ID() string {
	return s.id
}

// UserLogin returns the user login of the secret.
func (s *Secret) UserLogin() string {
	return s.userLogin
}

// Metadata returns the metadata of the secret.
func (s *Secret) Metadata() map[string]string {
	return s.metadata
}

// Data returns the data of the secret.
func (s *Secret) Data() *Data {
	return s.data
}

// Key returns the key of the secret.
func (d *Data) Key() string {
	return d.key
}

// Value returns the value of the secret.
func (d *Data) Value() string {
	return d.value
}
