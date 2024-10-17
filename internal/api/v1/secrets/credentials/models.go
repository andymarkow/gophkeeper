package credentials

import "time"

// Credential represents credentials.
type Credential struct {
	// ID represents credential id.
	ID string `json:"id"`

	// Metadata represents credential metadata.
	Metadata map[string]string `json:"metadata,omitempty"`

	// CreatedAt represents credential creation time.
	CreatedAt time.Time `json:"created_at,omitempty"`

	// UpdatedAt represents credential update time.
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// Data represents credential data.
	Data *Data `json:"data,omitempty"`
}

// Data represents credentials data.
type Data struct {
	// Login represents credential login.
	Login string `json:"login"`

	// Password represents credential password.
	Password string `json:"password"`
}

// CreateCredentialRequest represents a request to create a new credential.
type CreateCredentialRequest struct {
	*Credential
}

// CreateCredentialResponse represents a response to create a new credential.
type CreateCredentialResponse struct {
	Message string `json:"message"`
}

// ListCredentialsResponse represents a response to list credentials.
type ListCredentialsResponse struct {
	Creds []*Credential `json:"credentials"`
}

// GetCredentialResponse represents a response to get credential.
type GetCredentialResponse struct {
	Credential
}

// UpdateCredentialRequest represents update bank card request.
type UpdateCredentialRequest struct {
	Credential
}
