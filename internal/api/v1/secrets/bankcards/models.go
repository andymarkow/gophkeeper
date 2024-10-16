//nolint:tagliatelle
package bankcards

type BankCard struct {
	// ID represents bank card ID.
	ID string `json:"id"`

	// Metadata represents bank card metadata.
	Metadata map[string]string `json:"metadata,omitempty"`

	// Data represents bank card data.
	Data *Data `json:"data"`
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
	IDs []string `json:"ids"`
}

// UpdateCardRequest represents update bank card request.
type UpdateCardRequest struct {
	*BankCard
}
