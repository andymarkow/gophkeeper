package cardrepo

import (
	"context"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/bankcard"
)

type Storage interface {
	Add(ctx context.Context, card *bankcard.BankCard) error
	Get(ctx context.Context, userLogin string, cardID string) (*bankcard.BankCard, error)
	List(ctx context.Context, userLogin string) ([]*bankcard.BankCard, error)
	Update(ctx context.Context, card *bankcard.BankCard) error
	Delete(ctx context.Context, userLogin string, cardID string) error
}
