//nolint:tagliatelle
package bankcards

import "time"

type Secret struct {
	// ID represents secret ID.
	ID string `json:"id"`

	// Name represents secret name.
	Name string `json:"name"`

	// UserID represents secret user ID.
	UserID string `json:"user_id,omitempty"`

	// Metadata represents secret metadata.
	Metadata map[string]string `json:"metadata,omitempty"`

	// CreatedAt represents secret create at.
	CreatedAt time.Time `json:"created_at,omitempty"`

	// UpdatedAt represents secret update at.
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// Version represents secret version.
	Version int `json:"version,omitempty"`

	// Data represents secret data.
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
