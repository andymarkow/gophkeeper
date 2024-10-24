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
	"github.com/andymarkow/gophkeeper/internal/services/filesvc"
	"github.com/andymarkow/gophkeeper/internal/services/textsvc"
	"github.com/andymarkow/gophkeeper/internal/slogger"
	"github.com/andymarkow/gophkeeper/internal/storage/cardrepo"
	"github.com/andymarkow/gophkeeper/internal/storage/cardrepo/cardinmem"
	"github.com/andymarkow/gophkeeper/internal/storage/cardrepo/cardpg"
	"github.com/andymarkow/gophkeeper/internal/storage/credrepo"
	"github.com/andymarkow/gophkeeper/internal/storage/credrepo/credinmem"
	"github.com/andymarkow/gophkeeper/internal/storage/credrepo/credpg"
	"github.com/andymarkow/gophkeeper/internal/storage/filerepo"
	"github.com/andymarkow/gophkeeper/internal/storage/filerepo/fileinmem"
	"github.com/andymarkow/gophkeeper/internal/storage/filerepo/filepg"
	"github.com/andymarkow/gophkeeper/internal/storage/objrepo"
	"github.com/andymarkow/gophkeeper/internal/storage/pgmigrate"
	"github.com/andymarkow/gophkeeper/internal/storage/textrepo"
	"github.com/andymarkow/gophkeeper/internal/storage/userrepo"
	"github.com/andymarkow/gophkeeper/internal/storage/userrepo/userinmem"
	"github.com/andymarkow/gophkeeper/internal/storage/userrepo/userpg"
)

type Server struct {
	log     *slog.Logger
	httpsrv *httpserver.HTTPServer

	userStorage userrepo.Storage
	cardStorage cardrepo.Storage
	credStorage credrepo.Storage
	fileSvc     filesvc.Service
	textSvc     textsvc.Service
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

	objStorage, err := objrepo.NewMinioClient(cfg.ObjStorage.Endpoint, cfg.ObjStorage.Bucket, &objrepo.MinioClientOpts{
		AccessKeyID:     cfg.ObjStorage.AccessKey,
		SecretAccessKey: cfg.ObjStorage.SecretKey,
		UseSSL:          cfg.ObjStorage.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("objrepo.NewMinioClient: %w", err)
	}

	if err := objStorage.InitBucket(context.Background()); err != nil {
		return nil, fmt.Errorf("objStorage.InitBucket: %w", err)
	}

	textSvc := textsvc.NewSecretService(
		textrepo.NewInMemory(),
		objStorage,
		textsvc.WithLogger(logger),
		textsvc.WithCryptoKey([]byte(cfg.CryptoKey)),
		textsvc.WithObjectBasePath("texts"),
	)

	fileSvc, err := createFileSvc(cfg.Database.DSN, cfg.CryptoKey, logger, objStorage)
	if err != nil {
		return nil, fmt.Errorf("failed to init file service: %w", err)
	}

	if cfg.Database.DSN != "" {
		err := pgmigrate.Run(context.Background(), cfg.Database.DSN)
		if err != nil {
			return nil, fmt.Errorf("failed to apply migrations: %w", err)
		}
	}

	userStorage, err := createUserRepo(cfg.Database.DSN, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init user storage: %w", err)
	}

	cardStorage, err := createCardRepo(cfg.Database.DSN, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init card storage: %w", err)
	}

	credStorage, err := createCredRepo(cfg.Database.DSN, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init cred storage: %w", err)
	}

	router := router.NewRouter(
		userStorage,
		cardStorage,
		credStorage,
		textSvc,
		fileSvc,
		router.WithJWTSecret([]byte(cfg.JWTSecret)),
		router.WithCryptoKey([]byte(cfg.CryptoKey)),
		router.WithLogger(logger),
	)

	return &Server{
		log:         logger,
		httpsrv:     httpserver.NewHTTPServer(router, httpserver.WithLogger(logger)),
		userStorage: userStorage,
		cardStorage: cardStorage,
		credStorage: credStorage,
		fileSvc:     fileSvc,
		textSvc:     textSvc,
	}, nil
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
	signal.Notify(quit, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

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

	s.close()

	return nil
}

func (s *Server) close() {
	if err := s.userStorage.Close(); err != nil {
		s.log.Error("failed to close user storage", slog.Any("error", err))
	}

	if err := s.cardStorage.Close(); err != nil {
		s.log.Error("failed to close bank cards storage", slog.Any("error", err))
	}

	if err := s.credStorage.Close(); err != nil {
		s.log.Error("failed to close credentials storage", slog.Any("error", err))
	}

	if err := s.fileSvc.Close(); err != nil {
		s.log.Error("failed to close file service", slog.Any("error", err))
	}

	// if err := s.textSvc.Close(); err != nil {
	// 	s.log.Error("failed to close text service", slog.Any("error", err))
	// }
}

func createUserRepo(connStr string, logger *slog.Logger) (userrepo.Storage, error) {
	if connStr == "" {
		return userinmem.NewInMemory(), nil
	}

	repo, err := userpg.NewStorage(connStr, userpg.WithLogger(logger))
	if err != nil {
		return nil, fmt.Errorf("userpg.NewStorage: %w", err)
	}

	return repo, nil
}

func createCardRepo(connStr string, logger *slog.Logger) (cardrepo.Storage, error) {
	if connStr == "" {
		return cardinmem.NewInMemory(), nil
	}

	repo, err := cardpg.NewStorage(connStr, cardpg.WithLogger(logger))
	if err != nil {
		return nil, fmt.Errorf("cardpg.NewStorage: %w", err)
	}

	return repo, nil
}

func createCredRepo(connStr string, logger *slog.Logger) (credrepo.Storage, error) {
	if connStr == "" {
		return credinmem.NewInMemory(), nil
	}

	repo, err := credpg.NewStorage(connStr, credpg.WithLogger(logger))
	if err != nil {
		return nil, fmt.Errorf("credpg.NewStorage: %w", err)
	}

	return repo, nil
}

func createFileSvc(connStr, cryptoKey string, logger *slog.Logger, objStorage objrepo.Storage) (filesvc.Service, error) {
	var storage filerepo.Storage

	if connStr == "" {
		storage = fileinmem.NewInMemory()
	} else {
		var err error

		storage, err = filepg.NewStorage(connStr, filepg.WithLogger(logger))
		if err != nil {
			return nil, fmt.Errorf("filepg.NewStorage: %w", err)
		}
	}

	svc := filesvc.NewFileService(
		storage,
		objStorage,
		filesvc.WithLogger(logger),
		filesvc.WithCryptoKey([]byte(cryptoKey)),
		filesvc.WithObjectBasePath("files"),
	)

	return svc, nil
}
