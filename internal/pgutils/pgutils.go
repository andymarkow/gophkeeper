// Package pgutils provides PostgreSQL utilities.
package pgutils

import (
	"errors"
	"fmt"
	"syscall"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// isRetryableError checks if the error is retryable.
func isRetryableError(err error) bool {
	// Connection refused error
	if errors.Is(err, syscall.ECONNREFUSED) {
		return true
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
		// https://github.com/jackc/pgerrcode/blob/6e2875d9b438d43808cc033afe2d978db3b9c9e7/errcode.go#L393C6-L393C27
		return true
	}

	return false
}

// WithRetry retries operation in case of retryable errors.
func WithRetry(operation func() error) error {
	// Retry count
	retryCount := 3

	// Initial retry wait time
	var retryWaitTime time.Duration

	// Define the interval between retries
	retryWaitInterval := 2

	var err error

	for i := range retryCount {
		err = operation()
		if err == nil {
			return nil
		}

		if isRetryableError(err) {
			retryWaitTime = time.Duration((i*retryWaitInterval + 1)) * time.Second // 1s, 3s, 5s, etc.

			time.Sleep(retryWaitTime)
		} else {
			return fmt.Errorf("%w", err)
		}
	}

	return fmt.Errorf("retry attempts exceeded: %w", err)
}
