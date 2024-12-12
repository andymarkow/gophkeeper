package texts

import (
	"errors"
	"net/http"

	"github.com/andymarkow/gophkeeper/internal/httperr"
)

var (
	ErrSecretNameEmpty = httperr.NewHTTPError(http.StatusBadRequest, errors.New("secret name is empty"))
)
