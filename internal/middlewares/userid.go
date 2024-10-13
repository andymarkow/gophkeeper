package middlewares

import (
	"errors"
	"net/http"

	"github.com/go-chi/jwtauth/v5"

	"github.com/andymarkow/gophkeeper/internal/httperr"
)

// UserID is a middleware that fetches the user login from the JWT subject claim
// and sets it in the request header "X-User-Id".
func UserID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		token, _, err := jwtauth.FromContext(req.Context())
		if err != nil {
			httperr.HandleError(w, httperr.NewHTTPError(http.StatusInternalServerError, err))

			return
		}

		// Fetch the user login from the JWT sub claim field.
		userLogin := token.Subject()

		if userLogin == "" {
			httperr.HandleError(w, httperr.NewHTTPError(http.StatusUnauthorized, errors.New("user login is empty")))

			return
		}

		req.Header.Set("X-User-Id", userLogin)

		next.ServeHTTP(w, req)
	})
}