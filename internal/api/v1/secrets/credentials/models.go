package credentials

import "time"

// Secret represents a credential secret.
//
//nolint:tagliatelle
type Secret struct {
	// ID represents credential id.
	ID string `json:"id"`

	// Name represents credential name.
	Name string `json:"name"`

	// Metadata represents credential metadata.
	Metadata map[string]string `json:"metadata,omitempty"`

	// CreatedAt represents credential creation time.
	CreatedAt time.Time `json:"created_at,omitempty"`

	// UpdatedAt represents credential update time.
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// Data represents credential data.
	Data *Data `json:"data,omitempty"`
}

// Data represents credentials secret data.
type Data struct {
	// Login represents credential login.
	Login string `json:"login"`

	// Password represents credential password.
	Password string `json:"password"`
}

// CreateSecretRequest represents a request to create a new credential.
type CreateSecretRequest struct {
	Secret
}

// ListSecretsResponse represents a response to list credentials.
type ListSecretsResponse struct {
	Secrets []*Secret `json:"secrets"`
}

// GetSecretResponse represents a response to get credential.
type GetSecretResponse struct {
	*Secret
}

// UpdateSecretRequest represents update bank card request.
type UpdateSecretRequest struct {
	Secret
}
