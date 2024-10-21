package files

import "time"

// Secret represents a file secret.
//
//nolint:tagliatelle
type Secret struct {
	ID        string            `json:"id,omitempty"`
	Name      string            `json:"name,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at,omitempty"`
	UpdatedAt time.Time         `json:"updated_at,omitempty"`
	File      *File             `json:"file,omitempty"`
}

// File represents a file info.
type File struct {
	Name     string `json:"name,omitempty"`
	Checksum string `json:"checksum,omitempty"`
	Size     int64  `json:"size,omitempty"`
}

// CreateSecretRequest represents a request for creating a file.
type CreateSecretRequest struct {
	Name     string            `json:"name"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// CreateSecretResponse represents a response for creating a file secret.
type CreateSecretResponse struct {
	*Secret
}

// UpdateSecretRequest represents a request for updating a file secret.
type UpdateSecretRequest struct {
	Metadata map[string]string `json:"metadata,omitempty"`
	File     File              `json:"file,omitempty"`
}

// UpdateSecretResponse represents a response for updating a file secret.
type UpdateSecretResponse struct {
	*Secret
}

// ListSecretsResponse represents a response for listing file secrets.
type ListSecretsResponse struct {
	Secrets []*Secret `json:"secrets"`
}
