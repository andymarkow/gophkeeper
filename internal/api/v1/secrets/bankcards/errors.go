package bankcards

import (
	"errors"
	"net/http"

	"github.com/andymarkow/gophkeeper/internal/httperr"
)

var (
	ErrReqPayloadEmpty  = httperr.NewHTTPError(http.StatusBadRequest, errors.New("request payload is empty"))
	ErrUsrIDHeaderEmpty = httperr.NewHTTPError(http.StatusBadRequest, errors.New("X-User-Id header is empty"))
	ErrCardIDEmpty      = httperr.NewHTTPError(http.StatusBadRequest, errors.New("card ID is empty"))
)
