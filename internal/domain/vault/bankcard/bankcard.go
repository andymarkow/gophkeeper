// Package bankcard provides the domain model for bank cards.
package bankcard

import (
	"fmt"
)

// BankCard represents bank card.
type BankCard struct {
	id        string
	userLogin string
	metadata  map[string]string
	data      *Data
}

// Data represents bank card data.
type Data struct {
	number   string
	name     string
	cvv      string
	expireAt string
}

// NewBankCard creates a new bank card.
func NewBankCard(id, userLogin string, metadata map[string]string, data *Data) (*BankCard, error) {
	if id == "" {
		return nil, fmt.Errorf("id must not be empty")
	}

	if userLogin == "" {
		return nil, fmt.Errorf("user login must not be empty")
	}

	if data.number == "" {
		return nil, fmt.Errorf("number must not be empty")
	}

	if data.name == "" {
		return nil, fmt.Errorf("name must not be empty")
	}

	if data.cvv == "" {
		return nil, fmt.Errorf("cvv must not be empty")
	}

	if data.expireAt == "" {
		return nil, fmt.Errorf("expire at must not be empty")
	}

	return &BankCard{
		id:        id,
		userLogin: userLogin,
		metadata:  metadata,
		data:      data,
	}, nil
}

// ID returns the id of the bank card.
func (c *BankCard) ID() string {
	return c.id
}

// UserLogin returns the user login of the bank card.
func (c *BankCard) UserLogin() string {
	return c.userLogin
}

// Metadata returns the metadata of the bank card.
func (c *BankCard) Metadata() map[string]string {
	return c.metadata
}

// Data returns the data of the bank card.
func (c *BankCard) Data() *Data {
	return c.data
}

// Number returns the number of the bank card.
func (d *Data) Number() string {
	return d.number
}

// Name returns the name of the bank card.
func (d *Data) Name() string {
	return d.name
}

// CVV returns the CVV of the bank card.
func (d *Data) CVV() string {
	return d.cvv
}

// ExpireAt returns the expire at of the bank card.
func (d *Data) ExpireAt() string {
	return d.expireAt
}

func (c *BankCard) Encrypt() (*BankCard, error) {
	// TODO: implement encryption
	return nil, nil
}

func (c *BankCard) Decrypt() (*BankCard, error) {
	// TODO: implement decryption
	return nil, nil
}
