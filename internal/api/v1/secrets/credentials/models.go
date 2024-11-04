package credentials

import "time"

// Secret represents a credential secret.
//
//nolint:tagliatelle
type Secret struct {
	// ID represents secret id.
	ID string `json:"id"`

	// Name represents secret name.
	Name string `json:"name"`

	// UserID represents secret user ID.
	UserID string `json:"user_id,omitempty"`

	// Metadata represents secret metadata.
	Metadata map[string]string `json:"metadata,omitempty"`

	// CreatedAt represents secret creation time.
	CreatedAt time.Time `json:"created_at,omitempty"`

	// UpdatedAt represents secret update time.
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// Version represents secret version.
	Version int `json:"version,omitempty"`

	// Data represents secret data.
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
