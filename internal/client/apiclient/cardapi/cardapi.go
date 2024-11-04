// Package cardapi provides the API client.
package cardapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-resty/resty/v2"

	cache "github.com/andymarkow/gophkeeper/internal/cache/bankcard"
	"github.com/andymarkow/gophkeeper/internal/client/apierr"
	"github.com/andymarkow/gophkeeper/internal/domain/vault/bankcard"
)

type Client struct {
	log    *slog.Logger
	client *resty.Client
	cache  cache.Cache
	ttl    time.Duration
}

func NewClient(cache cache.Cache, opts ...Option) *Client {
	client := &Client{
		log:    slog.Default(),
		client: resty.New(),
		cache:  cache,
		ttl:    time.Second * 600,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

type Option func(c *Client)

func WithLogger(l *slog.Logger) Option {
	return func(c *Client) {
		c.log = l
	}
}

func WithCacheTTL(ttl time.Duration) Option {
	return func(c *Client) {
		c.ttl = ttl
	}
}

// WithBaseURL sets the base URL for the user API client.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.client.SetBaseURL(url)
	}
}

func WithAuthToken(token string) Option {
	return func(c *Client) {
		c.client.SetAuthToken(token)
	}
}

func (c *Client) CreateSecret(ctx context.Context, secret *bankcard.Secret) error {
	_, err := c.updateCacheItem(ctx, secret)
	if err != nil {
		c.log.Error("failed to update cache item", slog.Any("error", err))
	}

	_, err = c.DoCreateSecret(ctx, secret)
	if err != nil {
		if errors.Is(err, ErrSecretAlreadyExists) {
			return ErrSecretAlreadyExists
		}

		return fmt.Errorf("failed to create secret: %w", err)
	}

	return nil
}

func (c *Client) GetSecret(ctx context.Context, userID, secretName string) (*bankcard.Secret, error) {
	item, err := c.cache.GetItem(ctx, userID, secretName)
	if err != nil && errors.Is(err, cache.ErrItemNotFound) {
		secret, err := c.processGetSecret(ctx, secretName)
		if err != nil {
			if errors.Is(err, ErrSecretNotFound) {
				return nil, ErrSecretNotFound
			}

			return nil, fmt.Errorf("failed to get secret: %w", err)
		}

		return secret, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cache.GetItem: %w", err)
	}

	if item.Expired() {
		secret, err := c.processGetSecret(ctx, secretName)
		if err != nil {
			if errors.Is(err, ErrSecretNotFound) {
				return nil, ErrSecretNotFound
			}

			if apierr.IsConnError(err) {
				return item.Secret(), nil
			}

			return nil, fmt.Errorf("failed to get secret: %w", err)
		}

		return secret, nil
	}

	return item.Secret(), nil
}

func (c *Client) processGetSecret(ctx context.Context, secretName string) (*bankcard.Secret, error) {
	secret, err := c.DoGetSecret(ctx, secretName)
	if err != nil {
		if errors.Is(err, ErrSecretNotFound) {
			return nil, ErrSecretNotFound
		}

		return nil, fmt.Errorf("client.DoGetSecret: %w", err)
	}

	_, err = c.updateCacheItem(ctx, secret)
	if err != nil {
		c.log.Error("failed to update cache item", slog.Any("error", err))
	}

	return secret, nil
}

func (c *Client) ListSecrets(ctx context.Context, userID string) ([]*bankcard.Secret, error) {
	secrets, err := c.DoListSecrets(ctx)
	if err != nil {
		if apierr.IsConnError(err) {
			items, err := c.cache.ListItems(ctx, userID)
			if err != nil {
				return nil, fmt.Errorf("cache.ListItems: %w", err)
			}

			var secrets []*bankcard.Secret

			for _, item := range items {
				secrets = append(secrets, item.Secret())
			}

			return secrets, nil
		}

		return nil, fmt.Errorf("client.DoListSecrets: %w", err)
	}

	return secrets, nil
}

func (c *Client) UpdateSecret(ctx context.Context, secret *bankcard.Secret) error {
	updSecret, err := c.DoUpdateSecret(ctx, secret)
	if err != nil {
		if errors.Is(err, ErrSecretNotFound) {
			return ErrSecretNotFound
		}

		return fmt.Errorf("failed to update secret: %w", err)
	}

	_, err = c.updateCacheItem(ctx, updSecret)
	if err != nil {
		c.log.Error("failed to update cache item", slog.Any("error", err))
	}

	return nil
}

func (c *Client) DeleteSecret(ctx context.Context, userID, secretName string) error {
	defer func() {
		err := c.cache.DeleteItem(ctx, userID, secretName)
		if err != nil {
			if errors.Is(err, cache.ErrItemNotFound) {
				c.log.Error("cache item not found", slog.String("secret", secretName))

				return
			}

			c.log.Error("failed to delete cache item", slog.Any("error", err))
		}
	}()

	err := c.DoDeleteSecret(ctx, secretName)
	if err != nil {
		if errors.Is(err, ErrSecretNotFound) {
			return ErrSecretNotFound
		}

		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}

func (c *Client) updateCacheItem(ctx context.Context, secret *bankcard.Secret) (*cache.Item, error) {
	item := cache.NewItem(secret, time.Now().Add(c.ttl))

	it, err := c.cache.SetItem(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("cache.SetItem: %w", err)
	}

	return it, nil
}
