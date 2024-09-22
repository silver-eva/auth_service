package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"

	"github.com/silver-eva/auth_service/auth_service/config"
	"github.com/silver-eva/auth_service/auth_service/db"
	"github.com/silver-eva/auth_service/auth_service/lib/handlers"
	"github.com/silver-eva/auth_service/auth_service/lib/logger"
)

// var db db.PostgresInterface

func main() {
	if len(os.Args) > 2 {
		fmt.Println("Only one argument is allowed.")
		os.Exit(1)
	}

	configPath := ""
	if len(os.Args) == 2 {
		configPath = os.Args[1]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := logger.NewLogger(cfg.Env)
	db, err := db.New(cfg.StorageLink)
	if err != nil {
		logger.Error("Failed to create storage", slog.Any("error", err))
		return
	}

	router := http.NewServeMux()

	router.HandleFunc("POST /signup", handlers.SignupHandler(db, logger))
	router.HandleFunc("POST /login", handlers.LoginHandler(db, logger))
	router.HandleFunc("POST /auth", handlers.AuthHandler(db, logger))
	router.HandleFunc("POST /logout", handlers.LogoutHandler(db, logger))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.HTTPServer.Host, cfg.HTTPServer.Port),
		Handler: router,

		IdleTimeout:  cfg.IdleTimeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				logger.Info("Manual interruption", slog.String("host", server.Addr))
			} else {
				logger.Error("Server error", slog.Any("error", err))
			}
		}
	}()

	logger.Info("Server started", slog.String("host", server.Addr))

	<-done
	logger.Info("Server stopping", slog.String("host", server.Addr))

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTPServer.IdleTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown server", slog.Any("error", err))
		return
	}
}
