package files

import (
	"errors"
	"net/http"

	"github.com/andymarkow/gophkeeper/internal/httperr"
)

var (
	ErrFileIDEmpty = httperr.NewHTTPError(http.StatusBadRequest, errors.New("file ID is empty"))
)
