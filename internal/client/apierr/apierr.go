// Package apierr provides error handling for the API client.
package apierr

import (
	"net"
	"os"
)

// IsConnError checks if the error is a connection error.
func IsConnError(err error) bool {
	if opErr, ok := err.(*net.OpError); ok && opErr.Op == "dial" {
		return true
	}

	if os.IsTimeout(err) {
		return true
	}

	return false
}
