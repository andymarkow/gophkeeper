// Package credential provides the domain model for credentials.
package credential

import (
	"fmt"
	"time"
)

// Credential represents credentials.
type Credential struct {
	id       string
	userID   string
	metadata map[string]string
	createAt time.Time
	data     *Data
}

// Data represents credentials data.
type Data struct {
	login    string
	password string
}

// NewCredential creates a new credential.
func NewCredential(id, userID string, metadata map[string]string, createAt time.Time, data *Data) (*Credential, error) {
	if id == "" {
		return nil, fmt.Errorf("id must not be empty")
	}

	if userID == "" {
		return nil, fmt.Errorf("user id must not be empty")
	}

	return &Credential{
		id:       id,
		userID:   userID,
		metadata: metadata,
		createAt: createAt,
		data:     data,
	}, nil
}

// ID returns the id of the credential.
func (c *Credential) ID() string {
	return c.id
}

// UserID returns the user login of the credential.
func (c *Credential) UserID() string {
	return c.userID
}

// Metadata returns the metadata of the credential.
func (c *Credential) Metadata() map[string]string {
	return c.metadata
}

// CreateAt returns the create at of the credential.
func (c *Credential) CreateAt() time.Time {
	return c.createAt
}

// Data returns the data of the credential.
func (c *Credential) Data() *Data {
	return c.data
}

// Login returns the login of the credential.
func (d *Data) Login() string {
	return d.login
}

// Password returns the password of the credential.
func (d *Data) Password() string {
	return d.password
}
