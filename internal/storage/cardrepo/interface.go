// Package cardrepo provides bank cards storage implementation.
package cardrepo

import (
	"context"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/bankcard"
)

type Storage interface {
	AddCard(ctx context.Context, card *bankcard.Bankcard) error
	GetCard(ctx context.Context, userLogin string, cardID string) (*bankcard.Bankcard, error)
	ListCards(ctx context.Context, userLogin string) ([]*bankcard.Bankcard, error)
	UpdateCard(ctx context.Context, card *bankcard.Bankcard) error
	DeleteCard(ctx context.Context, userLogin string, cardID string) error
}
