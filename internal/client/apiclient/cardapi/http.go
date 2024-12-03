package cardapi

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/andymarkow/gophkeeper/internal/api/v1/secrets/bankcards"
	"github.com/andymarkow/gophkeeper/internal/domain/vault/bankcard"
)

func (c *Client) DoCreateSecret(ctx context.Context, secret *bankcard.Secret) (*bankcard.Secret, error) {
	var result bankcards.Secret

	body := bankcards.Secret{
		Name:     secret.Name(),
		Metadata: secret.Metadata(),
		Data: &bankcards.Data{
			Number:   secret.Data().Number(),
			Name:     secret.Data().Name(),
			CVV:      secret.Data().CVV(),
			ExpireAt: secret.Data().ExpireAt(),
		},
	}

	resp, err := c.client.R().SetContext(ctx).
		SetBody(body).
		SetResult(&result).
		Post("/api/v1/secrets/bankcards")
	if err != nil {
		return nil, fmt.Errorf("client.R.Post: %w", err)
	}

	if resp.IsError() {
		if resp.StatusCode() == 409 {
			return nil, ErrSecretAlreadyExists
		}

		c.log.Error("failed to create secret", slog.Any("error", resp.Error()), slog.Any("isError", resp.IsError()))

		return nil, fmt.Errorf("resp.IsError: %v", resp.Error())
	}

	data, err := bankcard.NewData(result.Data.Number, result.Data.Name, result.Data.CVV, result.Data.ExpireAt)
	if err != nil {
		return nil, fmt.Errorf("bankcard.NewData: %w", err)
	}

	secr, err := bankcard.NewSecret(
		result.ID,
		result.Name,
		result.UserID,
		result.Metadata,
		result.CreatedAt,
		result.UpdatedAt,
		result.Version,
		data,
	)
	if err != nil {
		return nil, fmt.Errorf("bankcard.NewSecret: %w", err)
	}

	return secr, nil
}

func (c *Client) DoGetSecret(ctx context.Context, secretName string) (*bankcard.Secret, error) {
	var result bankcards.Secret

	resp, err := c.client.R().SetContext(ctx).
		SetPathParams(map[string]string{
			"secretName": secretName,
		}).
		SetResult(result).
		Get("/api/v1/secrets/bankcards/{secretName}")
	if err != nil {
		return nil, fmt.Errorf("client.R.Get: %w", err)
	}

	if resp.IsError() {
		if resp.StatusCode() == 404 {
			return nil, ErrSecretNotFound
		}

		return nil, fmt.Errorf("resp.IsError: %v", resp.Error())
	}

	data, err := bankcard.NewData(result.Data.Number, result.Data.Name, result.Data.CVV, result.Data.ExpireAt)
	if err != nil {
		return nil, fmt.Errorf("bankcard.NewData: %w", err)
	}

	secret, err := bankcard.NewSecret(
		result.ID,
		result.Name,
		result.UserID,
		result.Metadata,
		result.CreatedAt,
		result.UpdatedAt,
		result.Version,
		data,
	)
	if err != nil {
		return nil, fmt.Errorf("bankcard.NewSecret: %w", err)
	}

	return secret, nil
}

func (c *Client) DoListSecrets(ctx context.Context) ([]*bankcard.Secret, error) {
	var result []bankcards.Secret

	resp, err := c.client.R().SetContext(ctx).
		SetResult(&result).
		Get("/api/secrets/v1/bankcards")
	if err != nil {
		return nil, fmt.Errorf("client.R.Get: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("resp.IsError: %v", resp.Error())
	}

	secrets := make([]*bankcard.Secret, 0, len(result))
	for _, secret := range result {
		data, err := bankcard.NewData(secret.Data.Number, secret.Data.Name, secret.Data.CVV, secret.Data.ExpireAt)
		if err != nil {
			return nil, fmt.Errorf("bankcard.NewData: %w", err)
		}

		secr, err := bankcard.NewSecret(
			secret.ID,
			secret.Name,
			secret.UserID,
			secret.Metadata,
			secret.CreatedAt,
			secret.UpdatedAt,
			secret.Version,
			data,
		)
		if err != nil {
			return nil, fmt.Errorf("bankcard.NewSecret: %w", err)
		}

		secrets = append(secrets, secr)
	}

	return secrets, nil
}

func (c *Client) DoUpdateSecret(ctx context.Context, secret *bankcard.Secret) (*bankcard.Secret, error) {
	var result bankcards.Secret

	body := bankcards.Secret{
		Name:     secret.Name(),
		Metadata: secret.Metadata(),
		Data: &bankcards.Data{
			Number:   secret.Data().Number(),
			Name:     secret.Data().Name(),
			CVV:      secret.Data().CVV(),
			ExpireAt: secret.Data().ExpireAt(),
		},
	}

	resp, err := c.client.R().SetContext(ctx).
		SetBody(body).
		SetPathParams(map[string]string{
			"secretName": secret.Name(),
		}).
		SetResult(&result).
		Put("/api/v1/secrets/bankcards/{secretName}")
	if err != nil {
		return nil, fmt.Errorf("client.R.Put: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("resp.IsError: %v", resp.Error())
	}

	data, err := bankcard.NewData(result.Data.Number, result.Data.Name, result.Data.CVV, result.Data.ExpireAt)
	if err != nil {
		return nil, fmt.Errorf("bankcard.NewData: %w", err)
	}

	secret, err = bankcard.NewSecret(
		result.ID,
		result.Name,
		result.UserID,
		result.Metadata,
		result.CreatedAt,
		result.UpdatedAt,
		result.Version,
		data,
	)
	if err != nil {
		return nil, fmt.Errorf("bankcard.NewSecret: %w", err)
	}

	return secret, nil
}

func (c *Client) DoDeleteSecret(ctx context.Context, secretName string) error {
	resp, err := c.client.R().SetContext(ctx).
		SetPathParams(map[string]string{
			"secretName": secretName,
		}).
		Delete("/api/v1/secrets/bankcards/{secretName}")
	if err != nil {
		return fmt.Errorf("client.R.Delete: %w", err)
	}

	if resp.IsError() {
		if resp.StatusCode() == 404 {
			return ErrSecretNotFound
		}

		return fmt.Errorf("resp.IsError: %v", resp.Error())
	}

	return nil
}
