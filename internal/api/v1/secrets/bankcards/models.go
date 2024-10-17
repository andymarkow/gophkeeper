//nolint:tagliatelle
package bankcards

import "time"

type BankCard struct {
	// ID represents bank card ID.
	ID string `json:"id"`

	// Metadata represents bank card metadata.
	Metadata map[string]string `json:"metadata,omitempty"`

	// CreatedAt represents bank card create at.
	CreatedAt time.Time `json:"created_at,omitempty"`

	// UpdatedAt represents bank card update at.
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// Data represents bank card data.
	Data *Data `json:"data,omitempty"`
}

// Data represents bank card data.
type Data struct {
	// Number represents bank card number.
	Number string `json:"number"`

	// Name represents bank card owner name.
	Name string `json:"name"`

	// CVV represents bank card CVV.
	CVV string `json:"cvv"`

	// ExpireAt represents bank card expire at.
	ExpireAt string `json:"expire_at"`
}

// CreateCardRequest represents create bank card request.
type CreateCardRequest struct {
	*BankCard
}

// CreateCardResponse represents create bank card response.
type CreateCardResponse struct {
	// Message represents response message.
	Message string `json:"message"`
}

// GetCardResponse represents get bank card response.
type GetCardResponse struct {
	BankCard
}

// ListCardsResponse represents list bank cards response.
type ListCardsResponse struct {
	Cards []*BankCard `json:"cards"`
}

// UpdateCardRequest represents update bank card request.
type UpdateCardRequest struct {
	*BankCard
}
