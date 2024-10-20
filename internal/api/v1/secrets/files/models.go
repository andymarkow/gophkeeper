package files

import "time"

// File represents a file object.
//
//nolint:tagliatelle
type File struct {
	ID        string            `json:"id"`
	Name      string            `json:"name,omitempty"`
	Checksum  string            `json:"checksum,omitempty"`
	Size      int64             `json:"size,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at,omitempty"`
	UpdatedAt time.Time         `json:"updated_at,omitempty"`
}

// CreateFileRequest represents a request for creating a file.
type CreateFileRequest struct {
	*File
}

// CreateFileResponse represents a response for creating a file.
type CreateFileResponse struct {
	*File
}

// UpdateFileRequest represents a request for updating a file.
type UpdateFileRequest struct {
	Name     string            `json:"name,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// UpdateFileResponse represents a response for updating a file.
type UpdateFileResponse struct {
	*File
}

// ListFilesResponse represents a response for listing files.
type ListFilesResponse struct {
	Files []*File `json:"files"`
}

// UploadFileResponse represents a response for uploading a file.
type UploadFileResponse struct {
	*File
}
