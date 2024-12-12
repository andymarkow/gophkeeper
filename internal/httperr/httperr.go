// Package httperr provides HTTP errors.
package httperr

import (
	"encoding/json"
	"errors"
	"net/http"
)

var (
	ErrReqPayloadEmpty  = NewHTTPError(http.StatusBadRequest, errors.New("request payload is empty"))
	ErrUsrIDHeaderEmpty = NewHTTPError(http.StatusBadRequest, errors.New("X-User-Id header is empty"))
)

type HTTPError struct {
	code int
	msg  error
}

func NewHTTPError(code int, err error) *HTTPError {
	return &HTTPError{code: code, msg: err}
}

func (e *HTTPError) Code() int {
	return e.code
}

func (e *HTTPError) Error() string {
	return e.msg.Error()
}

type jsonResponse struct {
	Error string `json:"error"`
}

func HandleError(w http.ResponseWriter, err *HTTPError) {
	resp := &jsonResponse{
		Error: err.Error(),
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(err.Code())

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
