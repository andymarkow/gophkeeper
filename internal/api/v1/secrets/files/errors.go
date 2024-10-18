package files

import (
	"errors"
	"net/http"

	"github.com/andymarkow/gophkeeper/internal/httperr"
)

var (
	ErrResIDEmpty    = httperr.NewHTTPError(http.StatusBadRequest, errors.New("resource ID is empty"))
	ErrChecksumEmpty = httperr.NewHTTPError(http.StatusBadRequest, errors.New("checksum is empty"))
)
