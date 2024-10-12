// Package credential provides the domain model for credentials.
package credential

import "fmt"

// Credential represents credentials.
type Credential struct {
	id        string
	userLogin string
	metadata  map[string]string
	data      *Data
}

// Data represents credentials data.
type Data struct {
	login    string
	password string
}

// NewCredential creates a new credential.
func NewCredential(id, userLogin string, metadata map[string]string, data *Data) (*Credential, error) {
	if id == "" {
		return nil, fmt.Errorf("id must not be empty")
	}

	if userLogin == "" {
		return nil, fmt.Errorf("user login must not be empty")
	}

	if data.login == "" {
		return nil, fmt.Errorf("login must not be empty")
	}

	return &Credential{
		id:        id,
		userLogin: userLogin,
		metadata:  metadata,
		data:      data,
	}, nil
}

// ID returns the id of the credential.
func (c *Credential) ID() string {
	return c.id
}

// UserLogin returns the user login of the credential.
func (c *Credential) UserLogin() string {
	return c.userLogin
}

// Metadata returns the metadata of the credential.
func (c *Credential) Metadata() map[string]string {
	return c.metadata
}

// Data returns the data of the credential.
func (c *Credential) Data() *Data {
	return c.data
}

// Login returns the login of the credential.
func (d *Data) Login() string {
	return d.login
}

// Password returns the password of the credential.
func (d *Data) Password() string {
	return d.password
}
