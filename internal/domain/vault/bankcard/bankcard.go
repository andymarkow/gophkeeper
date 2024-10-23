// Package bankcard provides the domain model for bank cards.
package bankcard

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/andymarkow/gophkeeper/internal/cryptutils"
)

// Secret represents bank card secret.
type Secret struct {
	id        string
	name      string
	userID    string
	metadata  map[string]string
	createdAt time.Time
	updatedAt time.Time
	data      *Data
}

// Data represents bank card secret data.
type Data struct {
	number    string
	name      string
	cvv       string
	expiredAt string
}

func CreateData(number, name, cvv, expiredAt string) (*Data, error) {
	if number == "" {
		return nil, fmt.Errorf("card number must not be empty")
	}

	if err := validateCardCvv(cvv); err != nil {
		return nil, fmt.Errorf("invalid card cvv: %w", err)
	}

	if err := validateCardExpireAt(expiredAt); err != nil {
		return nil, fmt.Errorf("invalid card expiration date: %w", err)
	}

	return NewData(number, name, cvv, expiredAt)
}

func NewData(number, name, cvv, expiredAt string) (*Data, error) {
	if number == "" {
		return nil, fmt.Errorf("card number must not be empty")
	}

	if name == "" {
		return nil, fmt.Errorf("card name holder must not be empty")
	}

	if cvv == "" {
		return nil, fmt.Errorf("card cvv value must not be empty")
	}

	if expiredAt == "" {
		return nil, fmt.Errorf("card expiration date must not be empty")
	}

	return &Data{
		number:    number,
		name:      name,
		cvv:       cvv,
		expiredAt: expiredAt,
	}, nil
}

// NewEmptyData creates a new empty data.
func NewEmptyData() *Data {
	return &Data{}
}

// NewSecret creates a new bank card.
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

// CreateSecret creates a new bank card.
func CreateSecret(name, userID string, metadata map[string]string, data *Data) (*Secret, error) {
	if data == nil {
		data = NewEmptyData()
	}

	return NewSecret(uuid.New().String(), name, userID, metadata, time.Now(), time.Now(), data)
}

// ID returns the id of the bank card secret.
func (s *Secret) ID() string {
	return s.id
}

// Name returns the name of the bank card secret.
func (s *Secret) Name() string {
	return s.name
}

// UserID returns the user login of the bank card secret.
func (s *Secret) UserID() string {
	return s.userID
}

// Metadata returns the metadata of the bank card secret.
func (s *Secret) Metadata() map[string]string {
	return s.metadata
}

// AddMetadata adds metadata to the bank card secret.
func (s *Secret) AddMetadata(metadata map[string]string) {
	for k, v := range metadata {
		s.metadata[k] = v
	}
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

// CreatedAt returns the create at of the bank card secret.
func (s *Secret) CreatedAt() time.Time {
	return s.createdAt
}

// UpdatedAt returns the update at of the bank card secret.
func (s *Secret) UpdatedAt() time.Time {
	return s.updatedAt
}

// Data returns the data of the bank card secret.
func (s *Secret) Data() *Data {
	return s.data
}

// SetData sets the data of the bank card secret.
func (s *Secret) SetData(data *Data) {
	s.data = data
}

// DataJSON returns the data of the bank card secret.
func (s *Secret) DataJSON() ([]byte, error) {
	if s.data == nil {
		return nil, nil
	}

	dataMap := map[string]string{
		"number":     s.data.number,
		"name":       s.data.name,
		"cvv":        s.data.cvv,
		"expired_at": s.data.expiredAt,
	}

	data, err := json.Marshal(dataMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	return data, nil
}

// Number returns the number of the bank card secret.
func (d *Data) Number() string {
	return d.number
}

// Name returns the name of the bank card secret.
func (d *Data) Name() string {
	return d.name
}

// CVV returns the CVV of the bank card.
func (d *Data) CVV() string {
	return d.cvv
}

// ExpireAt returns the expire at of the bank card.
func (d *Data) ExpireAt() string {
	return d.expiredAt
}

// Encrypt encrypts bank card data with the given key.
func (d *Data) Encrypt(key []byte) (*Data, error) {
	number, err := cryptutils.EncryptString(key, d.number)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt card number: %w", err)
	}

	name, err := cryptutils.EncryptString(key, d.name)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt card holder name: %w", err)
	}

	cvv, err := cryptutils.EncryptString(key, d.cvv)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt card cvv value: %w", err)
	}

	expiredAt, err := cryptutils.EncryptString(key, d.expiredAt)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt card expire at date: %w", err)
	}

	return &Data{
		number:    number,
		name:      name,
		cvv:       cvv,
		expiredAt: expiredAt,
	}, nil
}

// Decrypt decrypts bank card data with the given key.
func (d *Data) Decrypt(key []byte) (*Data, error) {
	number, err := cryptutils.DecryptString(key, d.number)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt card number: %w", err)
	}

	name, err := cryptutils.DecryptString(key, d.name)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt card holder name: %w", err)
	}

	cvv, err := cryptutils.DecryptString(key, d.cvv)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt card cvv value: %w", err)
	}

	expiredAt, err := cryptutils.DecryptString(key, d.expiredAt)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt card expire at date: %w", err)
	}

	return &Data{
		number:    number,
		name:      name,
		cvv:       cvv,
		expiredAt: expiredAt,
	}, nil
}

// validateCardCvv checks the card CVV value is valid.
func validateCardCvv(cvv string) error {
	if len(cvv) != 3 {
		return fmt.Errorf("value is not 3 digits")
	}

	if _, err := strconv.Atoi(cvv); err != nil {
		return fmt.Errorf("value is not a number")
	}

	return nil
}

// validateCardExpireAt checks the card expiration date is valid.
func validateCardExpireAt(expiredAt string) error {
	_, err := time.Parse(time.RFC3339, expiredAt)
	if err != nil {
		return fmt.Errorf("cant parse date as RFC3339 format")
	}

	return nil
}

// UnmarshalData unmarshals bank card data.
func UnmarshalData(data []byte) (*Data, error) {
	var d map[string]string

	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	dat, err := NewData(d["number"], d["name"], d["cvv"], d["expired_at"])
	if err != nil {
		return nil, fmt.Errorf("failed to create data: %w", err)
	}

	return dat, nil
}
