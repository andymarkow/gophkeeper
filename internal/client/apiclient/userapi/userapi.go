// Package userapi provides the API client.
package userapi

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-resty/resty/v2"

	"github.com/andymarkow/gophkeeper/internal/api/v1/users"
)

// Client represents a user API client.
type Client struct {
	log    *slog.Logger
	client *resty.Client
}

// NewClient creates a new user API client.
func NewClient(opts ...Option) *Client {
	client := &Client{
		log:    slog.Default(),
		client: resty.New(),
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// Option represents a user API client option.
type Option func(c *Client)

// WithLogger sets the logger for the user API client.
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

// WithDebug sets the debug mode for the user API client.
func WithDebug(debug bool) Option {
	return func(c *Client) {
		c.client.SetDebug(debug)
	}
}

// DoSignUpUser sends a request to sign up a new user.
func (c *Client) DoSignUpUser(ctx context.Context, login, password string) (*SignUpUserResponse, error) {
	var result SignUpUserResponse

	req := users.SignUpUserRequest{
		Login:    login,
		Password: password,
	}

	resp, err := c.client.R().SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		Post("/api/v1/users/signup")
	if err != nil {
		return nil, fmt.Errorf("client.R.Post: %w", err)
	}

	if resp.IsError() {
		if resp.StatusCode() == 409 {
			return nil, ErrUserAlreadyExists
		}

		return nil, fmt.Errorf("resp.IsError: %v", resp.Error())
	}

	return &result, nil
}

// DoSignInUser sends a request to sign in an existing user.
func (c *Client) DoSignInUser(ctx context.Context, login, password string) (*SignInUserResponse, error) {
	var result SignInUserResponse

	req := users.SignInUserRequest{
		Login:    login,
		Password: password,
	}

	resp, err := c.client.R().SetContext(ctx).
		SetBody(req).
		SetResult(&result).
		Post("/api/v1/users/signin")
	if err != nil {
		return nil, fmt.Errorf("client.R.Post: %w", err)
	}

	if resp.IsError() {
		switch resp.StatusCode() {
		case 400:
			return nil, ErrUserInvalidCredentials
		case 404:
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("resp.IsError: %v", resp.Error())
	}

	return &result, nil
}
