//nolint:tagliatelle
package bankcards

import "time"

type Secret struct {
	// ID represents bank card ID.
	ID string `json:"id"`

	// Name represents bank card name.
	Name string `json:"name"`

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

// CreateSecretRequest represents create bank card secret request.
type CreateSecretRequest struct {
	Secret
}

// ListSecretsResponse represents list bank card secret entries response.
type ListSecretsResponse struct {
	Secrets []*Secret `json:"secrets"`
}

// UpdateSecretRequest represents update bank card secret request.
type UpdateSecretRequest struct {
	Secret
}
