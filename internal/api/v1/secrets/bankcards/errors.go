package bankcards

import (
	"errors"
	"net/http"

	"github.com/andymarkow/gophkeeper/internal/httperr"
)

var (
	ErrSecretNameEmpty = httperr.NewHTTPError(http.StatusBadRequest, errors.New("bank card secret name is empty"))
)
