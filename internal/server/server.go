// Package server provides the server.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/andymarkow/gophkeeper/internal/server/config"
	"github.com/andymarkow/gophkeeper/internal/server/httpserver"
	"github.com/andymarkow/gophkeeper/internal/server/router"
	"github.com/andymarkow/gophkeeper/internal/slogger"
	"github.com/andymarkow/gophkeeper/internal/storage/cardrepo"
	"github.com/andymarkow/gophkeeper/internal/storage/credrepo"
	"github.com/andymarkow/gophkeeper/internal/storage/userrepo"
)

type Server struct {
	log     *slog.Logger
	httpsrv *httpserver.HTTPServer
}

func NewServer() (*Server, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("config.NewConfig: %w", err)
	}

	logLevel, err := slogger.ParseLogLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("slogger.ParseLogLevel: %w", err)
	}

	logger := slogger.NewLogger(slogger.WithLevel(logLevel))

	router := router.NewRouter(
		userrepo.NewInMemory(),
		cardrepo.NewInMemory(),
		credrepo.NewInMemory(),
		[]byte(cfg.JWTSecret),
		[]byte(cfg.CryptoKey),
		router.WithLogger(logger),
	)

	server := &Server{
		log:     logger,
		httpsrv: httpserver.NewHTTPServer(router, httpserver.WithLogger(logger)),
	}

	return server, nil
}

func (s *Server) Run() error {
	errgrp, ctx := errgroup.WithContext(context.Background())

	errgrp.Go(func() error {
		if err := s.httpsrv.Serve(); err != nil {
			return fmt.Errorf("failed to start HTTP server: %w", err)
		}

		return nil
	})

	// Graceful shutdown handler.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-quit:
		s.log.Info("Interruption signal received")

		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.httpsrv.Shutdown(stopCtx); err != nil {
			return fmt.Errorf("httpsrv.Shutdown: %w", err)
		}

	case <-ctx.Done():
	}

	if err := errgrp.Wait(); err != nil {
		return fmt.Errorf("errgrp.Wait: %w", err)
	}

	return nil
}
