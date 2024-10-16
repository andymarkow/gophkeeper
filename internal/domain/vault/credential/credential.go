// Package credential provides the domain model for credentials.
package credential

import (
	"fmt"
	"time"

	"github.com/andymarkow/gophkeeper/internal/cryptutils"
)

// Credential represents credentials.
type Credential struct {
	id        string
	userID    string
	metadata  map[string]string
	createAt  time.Time
	updatedAt time.Time
	data      *Data
}

// Data represents credentials data.
type Data struct {
	login    string
	password string
}

func NewData(login, password string) (*Data, error) {
	if login == "" {
		return nil, fmt.Errorf("login must not be empty")
	}

	if password == "" {
		return nil, fmt.Errorf("password must not be empty")
	}

	return &Data{
		login:    login,
		password: password,
	}, nil
}

func CreateCredential(id, userID string, metadata map[string]string, data *Data) (*Credential, error) {
	return NewCredential(id, userID, metadata, time.Now(), time.Now(), data)
}

// NewCredential creates a new credential.
func NewCredential(id, userID string, metadata map[string]string, createAt, updateAt time.Time, data *Data) (*Credential, error) {
	if id == "" {
		return nil, fmt.Errorf("id must not be empty")
	}

	if userID == "" {
		return nil, fmt.Errorf("user id must not be empty")
	}

	if data == nil {
		return nil, fmt.Errorf("data must not be empty")
	}

	return &Credential{
		id:        id,
		userID:    userID,
		metadata:  metadata,
		createAt:  createAt,
		updatedAt: updateAt,
		data:      data,
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

// UpdatedAt returns the update at of the credential.
func (c *Credential) UpdatedAt() time.Time {
	return c.updatedAt
}

// SetData sets the data of the credential.
func (c *Credential) SetData(data *Data) {
	c.data = data
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

// Encrypt encrypts credential data with the given key.
func (d *Data) Encrypt(key []byte) (*Data, error) {
	login, err := cryptutils.EncryptString(d.login, key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt login: %w", err)
	}

	password, err := cryptutils.EncryptString(d.password, key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}

	return &Data{
		login:    login,
		password: password,
	}, nil
}

// Decrypt decrypts credential data with the given key.
func (d *Data) Decrypt(key []byte) (*Data, error) {
	login, err := cryptutils.DecryptString(d.login, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt login: %w", err)
	}

	password, err := cryptutils.DecryptString(d.password, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password: %w", err)
	}

	return &Data{
		login:    login,
		password: password,
	}, nil
}
