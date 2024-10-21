package texts

import "time"

// Secret represents a text secret.
//
//nolint:tagliatelle
type Secret struct {
	ID        string            `json:"id,omitempty"`
	Name      string            `json:"name,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at,omitempty"`
	UpdatedAt time.Time         `json:"updated_at,omitempty"`
	Content   *Content          `json:"content,omitempty"`
}

// Content represents a secret content info.
type Content struct {
	Checksum string `json:"checksum,omitempty"`
}

// CreateSecretRequest represents a request for creating a secret.
type CreateSecretRequest struct {
	Name     string            `json:"name"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ListSecretsResponse represents a response for listing secrets.
type ListSecretsResponse struct {
	Secrets []*Secret `json:"secrets"`
}

// UpdateSecretRequest represents a request for updating a secret.
type UpdateSecretRequest struct {
	Metadata map[string]string `json:"metadata,omitempty"`
}
