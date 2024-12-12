// Package cardapi provides the API client.
package cardapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-resty/resty/v2"

	"github.com/andymarkow/gophkeeper/internal/client/apierr"
	"github.com/andymarkow/gophkeeper/internal/domain/vault/bankcard"
	"github.com/andymarkow/gophkeeper/internal/storage/cardrepo"
)

type Client struct {
	log     *slog.Logger
	client  *resty.Client
	storage cardrepo.Storage
}

func NewClient(store cardrepo.Storage, opts ...Option) *Client {
	client := &Client{
		log:     slog.Default(),
		client:  resty.New(),
		storage: store,
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
	_, err := c.DoCreateSecret(ctx, secret)
	if err != nil {
		if errors.Is(err, ErrSecretAlreadyExists) {
			return ErrSecretAlreadyExists
		}

		return fmt.Errorf("failed to create secret: %w", err)
	}

	_, err = c.storage.AddSecret(ctx, secret)
	if err != nil {
		return fmt.Errorf("storage.AddSecret: %w", err)
	}

	return nil
}

func (c *Client) GetSecret(ctx context.Context, userID, secretName string) (*bankcard.Secret, error) {
	secret, err := c.DoGetSecret(ctx, secretName)
	if err != nil {
		if errors.Is(err, ErrSecretNotFound) {
			return nil, ErrSecretNotFound
		}

		if apierr.IsConnError(err) {
			secret, err := c.storage.GetSecret(ctx, userID, secretName)
			if err != nil {
				if errors.Is(err, cardrepo.ErrSecretNotFound) {
					return nil, ErrSecretNotFound
				}

				return nil, fmt.Errorf("storage.GetSecret: %w", err)
			}

			return secret, nil
		}

		return nil, fmt.Errorf("client.DoGetSecret: %w", err)
	}

	_, err = c.storage.UpdateSecret(ctx, secret)
	if err != nil {
		return nil, fmt.Errorf("storage.UpdateSecret: %w", err)
	}

	return secret, nil
}

func (c *Client) ListSecrets(ctx context.Context, userID string) ([]*bankcard.Secret, error) {
	secrets, err := c.DoListSecrets(ctx)
	if err != nil {
		if apierr.IsConnError(err) {
			secrets, err := c.storage.ListSecrets(ctx, userID)
			if err != nil {
				return nil, fmt.Errorf("cache.ListItems: %w", err)
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

	_, err = c.storage.UpdateSecret(ctx, updSecret)
	if err != nil {
		return fmt.Errorf("storage.UpdateSecret: %w", err)
	}

	return nil
}

func (c *Client) DeleteSecret(ctx context.Context, userID, secretName string) error {
	defer func() {
		err := c.storage.DeleteSecret(ctx, userID, secretName)
		if err != nil {
			if errors.Is(err, cardrepo.ErrSecretNotFound) {
				c.log.Error("secret not found", slog.String("secret", secretName))

				return
			}

			c.log.Error("failed to delete secret", slog.Any("error", err))
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
