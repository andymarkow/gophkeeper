// Package authconfig provides the auth config.
package authconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

//nolint:tagliatelle
type Content struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

type AuthConfig struct {
	userID   string
	username string
	token    string
}

func NewAuthConfig(userID, username, token string) *AuthConfig {
	return &AuthConfig{
		userID:   userID,
		username: username,
		token:    token,
	}
}

func (c *AuthConfig) UserID() string {
	return c.userID
}

func (c *AuthConfig) Username() string {
	return c.username
}

func (c *AuthConfig) Token() string {
	return c.token
}

func ReadFile(basepath string) (*AuthConfig, error) {
	filename := filepath.Join(basepath, "auth.json")

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var content Content

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&content); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	authcfg := NewAuthConfig(content.UserID, content.Username, content.Token)

	return authcfg, nil
}

func (c *AuthConfig) WriteFile(basepath string) error {
	filename := filepath.Join(basepath, "auth.json")

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	content := Content{
		UserID:   c.UserID(),
		Username: c.Username(),
		Token:    c.Token(),
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(content); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
