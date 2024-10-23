// Package userrepo provides storage implementation for user model.
package userrepo

import (
	"context"

	"github.com/andymarkow/gophkeeper/internal/domain/user"
)

// Storage represents user storage interface.
type Storage interface {
	AddUser(ctx context.Context, usr *user.User) error
	GetUser(ctx context.Context, login string) (*user.User, error)
}
