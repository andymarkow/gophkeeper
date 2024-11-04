// Package bankcard provides the cache implementation for bank cards.
package bankcard

import (
	"time"

	"github.com/andymarkow/gophkeeper/internal/domain/vault/bankcard"
)

type Item struct {
	secret    *bankcard.Secret
	expitedAt time.Time
}

func NewItem(secret *bankcard.Secret, expitedAt time.Time) *Item {
	return &Item{
		secret:    secret,
		expitedAt: expitedAt,
	}
}

func (i *Item) Expired() bool {
	return i.expitedAt.After(time.Now())
}

func (i *Item) Secret() *bankcard.Secret {
	return i.secret
}

func (i *Item) SetSecret(secret *bankcard.Secret) {
	i.secret = secret
}

func (i *Item) ExpitedAt() time.Time {
	return i.expitedAt
}

func (i *Item) SetExpiredAt(expitedAt time.Time) {
	i.expitedAt = expitedAt
}
