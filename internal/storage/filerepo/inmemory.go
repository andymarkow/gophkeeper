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
func (s *InMemory) AddFile(_ context.Context, file *fileobj.File) (*fileobj.File, error) {
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

		return file, nil
	}

	if _, ok := files[file.ID()]; ok {
		// File already exists in the storage.
		return nil, fmt.Errorf("%w: %s", ErrFileAlreadyExists, file.ID())
	}

	// Add file to the storage.
	s.files[file.UserID()][file.ID()] = *file

	return file, nil
}

// GetFile returns a file from the storage.
func (s *InMemory) GetFile(_ context.Context, userID, fileID string) (*fileobj.File, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the user login entry exists in the storage.
	files, ok := s.files[userID]
	if !ok {
		return nil, fmt.Errorf("%w for user login %s: %s", ErrFileNotFound, userID, fileID)
	}

	// Check if the file entry exists in the storage.
	if file, ok := files[fileID]; ok {
		f := file

		return &f, nil
	}

	return nil, fmt.Errorf("%w for user login %s: %s", ErrFileNotFound, userID, fileID)
}

// ListFiles returns a list of files from the storage.
func (s *InMemory) ListFiles(_ context.Context, userID string) ([]*fileobj.File, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	files := s.files[userID]

	fl := make([]*fileobj.File, 0, len(files))

	for _, file := range files {
		f := file
		fl = append(fl, &f)
	}

	return fl, nil
}

// UpdateFile updates a file in the storage.
func (s *InMemory) UpdateFile(_ context.Context, file *fileobj.File) (*fileobj.File, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user login entry exists in the storage.
	files, ok := s.files[file.UserID()]
	if !ok {
		return nil, fmt.Errorf("%w for user login %s: %s", ErrFileNotFound, file.UserID(), file.ID())
	}

	// Check if the file entry exists in the storage.
	if _, ok := files[file.ID()]; !ok {
		return nil, fmt.Errorf("%w for user login %s: %s", ErrFileNotFound, file.UserID(), file.ID())
	}

	// Update file in the storage.
	s.files[file.UserID()][file.ID()] = *file

	f := s.files[file.UserID()][file.ID()]

	return &f, nil
}

// DeleteFile deletes a file from the storage.
func (s *InMemory) DeleteFile(_ context.Context, userID, fileID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user login entry exists in the storage.
	files, ok := s.files[userID]
	if !ok {
		return fmt.Errorf("%w for user login %s: %s", ErrFileNotFound, userID, fileID)
	}

	// Check if the file entry exists in the storage.
	if _, ok := files[fileID]; !ok {
		return fmt.Errorf("%w for user login %s: %s", ErrFileNotFound, userID, fileID)
	}

	// Delete file from the storage.
	delete(s.files[userID], fileID)

	return nil
}
