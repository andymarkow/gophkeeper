package users

import (
	"errors"
	"net/http"

	"github.com/andymarkow/gophkeeper/internal/httperr"
)

//nolint:gochecknoglobals
var (
	ErrReqPayloadEmpty  = httperr.NewHTTPError(http.StatusBadRequest, errors.New("request payload is empty"))
	ErrUsrAlreadyExists = httperr.NewHTTPError(http.StatusConflict, errors.New("user already exists"))
	ErrUsrNotFound      = httperr.NewHTTPError(http.StatusNotFound, errors.New("user not found"))
	ErrUsrPassWrong     = httperr.NewHTTPError(http.StatusUnauthorized, errors.New("user password is wrong"))
)
