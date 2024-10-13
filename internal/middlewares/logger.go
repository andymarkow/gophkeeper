package middlewares

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

type responseData struct {
	status int
	size   int
}

// loggerResponseWriter wraps http.ResponseWriter and tracks the response size and status code.
// Uses in Logger middleware.
type loggerResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (w *loggerResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.responseData.size += size

	return size, err //nolint:wrapcheck
}

func (w *loggerResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.responseData.status = statusCode
}

// Logger is a router middleware that logs requests and their processing time.
func Logger(logg *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			startTime := time.Now()

			responseData := &responseData{
				status: 200,
				size:   0,
			}

			cw := loggerResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}

			remoteAddr, err := getRemoteIPAddr(req)
			if err != nil {
				logg.Error("failed to get remote address", slog.Any("error", err))
			}

			scheme := "http"

			if req.TLS != nil {
				scheme = "https"
			}

			defer func() {
				logg.Info("request",
					slog.String("remote_addr", remoteAddr.String()),
					slog.String("host", scheme+"://"+req.Host),
					slog.String("uri", req.RequestURI),
					slog.String("method", req.Method),
					slog.Int("status", responseData.status),
					slog.Int("size", responseData.size),
					slog.String("duration", time.Since(startTime).String()),
				)
			}()

			next.ServeHTTP(&cw, req)
		})
	}
}

// getRemoteIPAddr returns the remote IP address of the request. It first tries to
// parse the X-Real-IP header, then the X-Forwarded-For header, and finally the
// remote address. If any of these fail, it returns an error.
func getRemoteIPAddr(req *http.Request) (net.IP, error) {
	// Try the X-Real-IP header.
	ip := net.ParseIP(req.Header.Get("X-Real-IP"))
	if ip == nil {
		// If the X-Real-IP header is not set, try the X-Forwarded-For header.
		ips := strings.Split(req.Header.Get("X-Forwarded-For"), ",")

		if len(ips) > 0 {
			// Try to parse the first IP in the list.
			ip = net.ParseIP(ips[0])
		}
	}

	// If the IP is still not set, use the remote address.
	if ip == nil {
		ipStr, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse remote address: %w", err)
		}

		ip = net.ParseIP(ipStr)
	}

	return ip, nil
}
