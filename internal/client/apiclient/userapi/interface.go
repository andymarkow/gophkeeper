package userapi

import (
	"context"

	"github.com/andymarkow/gophkeeper/internal/api/v1/users"
)

type IClient interface {
	DoSignUpUser(ctx context.Context, login, password string) (*users.SignUpUserResponse, error)
	DoSignInUser(ctx context.Context, login, password string) (*users.SignInUserResponse, error)
}
