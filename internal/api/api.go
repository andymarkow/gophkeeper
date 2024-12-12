// Package api provides the API.
package api

import (
	"encoding/json"
	"net/http"
)

func JSONResponse(w http.ResponseWriter, status int, resp any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)

	if resp == nil {
		return
	}

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
