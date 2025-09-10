// @title			Backend API
// @version		1.0
// @description	API documentation for the blockchain backend payment system
// @host			localhost:8080
// @BasePath		/
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

	"backend/internal/database"
	"backend/internal/server"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	slog.Info("shutting down gracefully, press Ctrl+C again to force")
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown with error: ", slog.Any("error", err))
	}

	slog.Info("Server exiting")
	done <- true
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
	slog.SetDefault(logger)

	dbService := database.New("")
	defer func() {
		if err := dbService.Close(); err != nil {
			slog.Error("Failed to close database connection", slog.Any("error", err))
		}
	}()

	database.Migrate("")

	apiServer := server.NewServer(dbService)
	done := make(chan bool, 1)
	go gracefulShutdown(apiServer, done)

	slog.Info(fmt.Sprintf("Starting server on %s", apiServer.Addr))
	err := apiServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(fmt.Sprintf("http apiServer error: %s", err))
	}

	<-done
	slog.Info("Graceful shutdown complete.")
}