//nolint:tagliatelle
package bankcards

import "time"

// CreateCardRequest represents create bank card request.
type CreateCardRequest struct {
	// ID represents bank card ID.
	ID string `json:"id"`

	// Metadata represents bank card metadata.
	Metadata map[string]string `json:"metadata"`

	// Data represents bank card data.
	Data *CardData `json:"data"`
}

// CardData represents bank card data.
type CardData struct {
	// Number represents bank card number.
	Number string `json:"number"`

	// Name represents bank card owner name.
	Name string `json:"name"`

	// CVV represents bank card CVV.
	CVV string `json:"cvv"`

	// ExpireAt represents bank card expire at.
	ExpireAt time.Time `json:"expire_at"`
}

// CreateCardResponse represents create bank card response.
type CreateCardResponse struct {
	// Message represents response message.
	Message string `json:"message"`
}
