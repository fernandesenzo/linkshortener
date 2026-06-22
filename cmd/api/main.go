package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fernandesenzo/linkshortener/internal/infra"
	"github.com/fernandesenzo/linkshortener/internal/link/codegen"
	"github.com/fernandesenzo/linkshortener/internal/link/handler"
	"github.com/fernandesenzo/linkshortener/internal/link/repository"
	"github.com/fernandesenzo/linkshortener/internal/link/service"
	"github.com/fernandesenzo/linkshortener/internal/logger"
	"github.com/fernandesenzo/linkshortener/internal/middleware"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("main.run: could not read the .env file: %w", err)
	}

	logger.Setup()

	redisAddr := os.Getenv("REDIS_ADDR")
	redisPswd := os.Getenv("REDIS_PASSWORD")
	port := os.Getenv("SERVER_PORT")

	if redisAddr == "" || port == "" || redisPswd == "" {
		return fmt.Errorf("main.run: some variables from env came empty")
	}

	redisClient, err := infra.NewRedisClient(redisAddr, redisPswd)
	if err != nil {
		return fmt.Errorf("main.run: redis connection failed: %w", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			slog.Error("main.run: failed to close redis", "err", err)
		} else {
			slog.Info("redis connection closed gracefully")
		}
	}()
	slog.Info("connected to redis succesfully")

	codeGenerator := codegen.New()
	repo := repository.NewRedisRepository(redisClient)
	svc := service.New(codeGenerator, repo)
	h := handler.New(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /links", h.Create)
	mux.HandleFunc("GET /links/{id}", h.Get)

	var handlerStack http.Handler = mux
	handlerStack = middleware.AccessLog(handlerStack)
	handlerStack = middleware.ApplyHeaders(handlerStack)
	handlerStack = middleware.InjectReqID(handlerStack)
	handlerStack = middleware.Recover(handlerStack)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handlerStack,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)

	go func() {
		slog.Info("server starting", "port", port)
		serverErrors <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}
	case <-ctx.Done():
		slog.Info("shutting down OS signal received")

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelShutdown()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}
	}

	return nil
}
