package files

import "time"

// File represents a file object.
//
//nolint:tagliatelle
type File struct {
	ID        string            `json:"id"`
	Name      string            `json:"name,omitempty"`
	Checksum  string            `json:"checksum,omitempty"`
	Location  string            `json:"location,omitempty"`
	Size      int64             `json:"size,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at,omitempty"`
	UpdatedAt time.Time         `json:"updated_at,omitempty"`
}

type CreateFileResponse struct {
	*File
}
