package credentials

import (
	"errors"
	"net/http"

	"github.com/andymarkow/gophkeeper/internal/httperr"
)

var (
	ErrCredIDEmpty = httperr.NewHTTPError(http.StatusBadRequest, errors.New("credential ID is empty"))
)
