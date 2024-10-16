package bankcards

import (
	"errors"
	"net/http"

	"github.com/andymarkow/gophkeeper/internal/httperr"
)

var (
	ErrCardIDEmpty = httperr.NewHTTPError(http.StatusBadRequest, errors.New("card ID is empty"))
)
