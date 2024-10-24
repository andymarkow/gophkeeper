// Package credential provides the domain model for credentials.
package credential

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/andymarkow/gophkeeper/internal/cryptutils"
)

// Secret represents credential secret.
type Secret struct {
	id        string
	name      string
	userID    string
	metadata  map[string]string
	createdAt time.Time
	updatedAt time.Time
	data      *Data
}

// Data represents credential secret data.
type Data struct {
	login    string
	password string
}

// NewData creates a new data for the credential secret.
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

// NewEmptyData creates a new empty data for the credential secret.
func NewEmptyData() *Data {
	return &Data{}
}

// NewSecret creates a new credential.
func NewSecret(id, name, userID string, metadata map[string]string, createdAt, updatedAt time.Time, data *Data) (*Secret, error) {
	if id == "" {
		return nil, fmt.Errorf("id must not be empty")
	}

	if name == "" {
		return nil, fmt.Errorf("name must not be empty")
	}

	if userID == "" {
		return nil, fmt.Errorf("user id must not be empty")
	}

	if metadata == nil {
		metadata = make(map[string]string)
	}

	if data == nil {
		return nil, fmt.Errorf("data must not be empty")
	}

	return &Secret{
		id:        id,
		name:      name,
		userID:    userID,
		metadata:  metadata,
		createdAt: createdAt,
		updatedAt: updatedAt,
		data:      data,
	}, nil
}

// CreateSecret creates a new credential secret.
func CreateSecret(name, userID string, metadata map[string]string, data *Data) (*Secret, error) {
	if data == nil {
		data = NewEmptyData()
	}

	return NewSecret(uuid.New().String(), name, userID, metadata, time.Now(), time.Now(), data)
}

// ID returns the id of the credential secret.
func (s *Secret) ID() string {
	return s.id
}

// Name returns the name of the credential secret.
func (s *Secret) Name() string {
	return s.name
}

// UserID returns the user login of the credential secret.
func (s *Secret) UserID() string {
	return s.userID
}

// Metadata returns the metadata of the credential secret.
func (s *Secret) Metadata() map[string]string {
	return s.metadata
}

// AddMetadata adds metadata to the credential secret.
func (s *Secret) AddMetadata(metadata map[string]string) {
	s.metadata = metadata
}

// MetadataJSON returns the metadata of the bank card secret.
func (s *Secret) MetadataJSON() ([]byte, error) {
	if s.metadata == nil {
		return nil, nil
	}

	metadata, err := json.Marshal(s.metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return metadata, nil
}

// CreatedAt returns the create at of the credential secret.
func (s *Secret) CreatedAt() time.Time {
	return s.createdAt
}

// UpdatedAt returns the update at of the credential secret.
func (s *Secret) UpdatedAt() time.Time {
	return s.updatedAt
}

// Data returns the data of the credential.
func (s *Secret) Data() *Data {
	return s.data
}

// SetData sets the data of the credential secret.
func (s *Secret) SetData(data *Data) {
	s.data = data
}

// DataJSON returns the data of the bank card secret.
func (s *Secret) DataJSON() ([]byte, error) {
	if s.data == nil {
		return nil, nil
	}

	dataMap := map[string]string{
		"login":    s.data.login,
		"password": s.data.password,
	}

	data, err := json.Marshal(dataMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	return data, nil
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
	login, err := cryptutils.EncryptString(key, d.login)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt login: %w", err)
	}

	password, err := cryptutils.EncryptString(key, d.password)
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
	login, err := cryptutils.DecryptString(key, d.login)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt login: %w", err)
	}

	password, err := cryptutils.DecryptString(key, d.password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password: %w", err)
	}

	return &Data{
		login:    login,
		password: password,
	}, nil
}

// UnmarshalData unmarshals credential data.
func UnmarshalData(data []byte) (*Data, error) {
	var d map[string]string

	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	dat, err := NewData(d["login"], d["password"])
	if err != nil {
		return nil, fmt.Errorf("failed to create data: %w", err)
	}

	return dat, nil
}
