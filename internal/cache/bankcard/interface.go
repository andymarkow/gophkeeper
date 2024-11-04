package bankcard

import "context"

type Cache interface {
	SetItem(ctx context.Context, item *Item) (*Item, error)
	GetItem(ctx context.Context, userID, secretName string) (*Item, error)
	ListItems(ctx context.Context, userID string) ([]*Item, error)
	// UpdateItem(ctx context.Context, item *Item) (*Item, error)
	DeleteItem(ctx context.Context, userID, secretName string) error
}
