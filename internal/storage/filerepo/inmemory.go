package filerepo

import (
	"context"
	"fmt"
	"sync"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/fileobj"
)

// InMemory represents in-memory files storage.
type InMemory struct {
	// UserID -> Login -> File.
	files map[string]map[string]fileobj.File

	mu sync.RWMutex
}

// NewInMemory creates new in-memory files storage.
func NewInMemory() *InMemory {
	return &InMemory{
		files: make(map[string]map[string]fileobj.File),
	}
}

// AddFile adds a new file to the storage.
func (s *InMemory) AddFile(_ context.Context, file *fileobj.File) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user login entry exists in the storage.
	files, ok := s.files[file.UserID()]
	if !ok {
		// Check if files entry is nil.
		if files == nil {
			// Initialize files entry.
			s.files[file.UserID()] = make(map[string]fileobj.File)
		}

		// UserID does not exist in the storage. Add user login and file to the storage.
		s.files[file.UserID()][file.ID()] = *file

		return nil
	}

	if _, ok := files[file.ID()]; ok {
		// File already exists in the storage.
		return fmt.Errorf("%w: %s", ErrFileAlreadyExists, file.ID())
	}

	// Add file to the storage.
	s.files[file.UserID()][file.ID()] = *file

	return nil
}
