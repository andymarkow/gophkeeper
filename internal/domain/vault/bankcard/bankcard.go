// Package bankcard provides the domain model for bank cards.
package bankcard

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/andymarkow/gophkeeper/internal/cryptutils"
)

// BankCard represents bank card.
type BankCard struct {
	id       string
	userID   string
	metadata map[string]string
	createAt time.Time
	data     *Data
}

// Data represents bank card data.
type Data struct {
	number   string
	name     string
	cvv      string
	expireAt string
}

func CreateData(number, name, cvv string, expireAt time.Time) (*Data, error) {
	if !validateCardNumber(number) {
		return nil, fmt.Errorf("invalid card number")
	}

	if len(cvv) != 3 {
		return nil, fmt.Errorf("card cvv value is not 3 digits")
	}

	if _, err := strconv.Atoi(cvv); err != nil {
		return nil, fmt.Errorf("card cvv value is not a number")
	}

	if expireAt.IsZero() {
		return nil, fmt.Errorf("invalid card expiration date")
	}

	return NewData(number, name, cvv, expireAt.String())
}

func NewData(number, name, cvv, expireAt string) (*Data, error) {
	if number == "" {
		return nil, fmt.Errorf("card number must not be empty")
	}

	if name == "" {
		return nil, fmt.Errorf("card name holder must not be empty")
	}

	if cvv == "" {
		return nil, fmt.Errorf("card cvv value must not be empty")
	}

	if expireAt == "" {
		return nil, fmt.Errorf("card expiration date must not be empty")
	}

	return &Data{
		number:   number,
		name:     name,
		cvv:      cvv,
		expireAt: expireAt,
	}, nil
}

// CreateBankCard creates a new bank card.
func CreateBankCard(id, userID string, metadata map[string]string, data *Data) (*BankCard, error) {
	if data == nil {
		return nil, fmt.Errorf("data must not be empty")
	}

	return NewBankCard(id, userID, metadata, time.Now(), data)
}

// NewBankCard creates a new bank card.
func NewBankCard(id, userID string, metadata map[string]string, createAt time.Time, data *Data) (*BankCard, error) {
	if id == "" {
		return nil, fmt.Errorf("id must not be empty")
	}

	if userID == "" {
		return nil, fmt.Errorf("user id must not be empty")
	}

	if data == nil {
		return nil, fmt.Errorf("data must not be empty")
	}

	return &BankCard{
		id:       id,
		userID:   userID,
		metadata: metadata,
		createAt: createAt,
		data:     data,
	}, nil
}

// ID returns the id of the bank card.
func (c *BankCard) ID() string {
	return c.id
}

// UserID returns the user login of the bank card.
func (c *BankCard) UserID() string {
	return c.userID
}

// Metadata returns the metadata of the bank card.
func (c *BankCard) Metadata() map[string]string {
	return c.metadata
}

// CreateAt returns the create at of the bank card.
func (c *BankCard) CreateAt() time.Time {
	return c.createAt
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

// Encrypt encrypts bank card data with the given key.
func (d *Data) Encrypt(key []byte) (*Data, error) {
	number, err := cryptutils.EncryptString(d.number, key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt card number: %w", err)
	}

	name, err := cryptutils.EncryptString(d.name, key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt card holder name: %w", err)
	}

	cvv, err := cryptutils.EncryptString(d.cvv, key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt card cvv value: %w", err)
	}

	expireAt, err := cryptutils.EncryptString(d.expireAt, key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt card expire at date: %w", err)
	}

	return &Data{
		number:   number,
		name:     name,
		cvv:      cvv,
		expireAt: expireAt,
	}, nil
}

// Decrypt decrypts bank card data with the given key.
func (d *Data) Decrypt(key []byte) (*Data, error) {
	number, err := cryptutils.DecryptString(d.number, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt card number: %w", err)
	}

	name, err := cryptutils.DecryptString(d.name, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt card holder name: %w", err)
	}

	cvv, err := cryptutils.DecryptString(d.cvv, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt card cvv value: %w", err)
	}

	expireAt, err := cryptutils.DecryptString(d.expireAt, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt card expire at date: %w", err)
	}

	return &Data{
		number:   number,
		name:     name,
		cvv:      cvv,
		expireAt: expireAt,
	}, nil
}

// validateCardNumber checks id is valid or not based on Luhn algorithm.
func validateCardNumber(cardNumber string) bool {
	cardNumber = strings.ReplaceAll(cardNumber, " ", "")

	var sum int
	double := false

	for i := len(cardNumber) - 1; i >= 0; i-- {
		n := cardNumber[i]

		if !unicode.IsDigit(rune(n)) {
			return false // invalid character
		}

		digit := int(n - '0')
		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		double = !double
	}

	return sum%10 == 0
}
