package tests

import (
	"context"
	"log/slog"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func MustStartPostgresContainer() (func(context.Context, ...testcontainers.TerminateOption) error, string, error) {
	dbName := "database"
	dbUser := "user"
	dbPwd := "password"

	ctx := context.Background()
	dbContainer, err := postgres.Run(
		ctx,
		"postgres:latest",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPwd),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10*time.Second)),
	)
	if err != nil {
		return nil, "", err
	}

	// Build connection string (sslmode=disable to keep things simple)
	connStr, err := dbContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return dbContainer.Terminate, "", err
	}

	return dbContainer.Terminate, connStr, nil
}

func Cleanup(teardown func(context context.Context, opts ...testcontainers.TerminateOption) error) {
	if teardown != nil {
		if err := teardown(context.Background()); err != nil {
			slog.Error("could not teardown postgres container:", slog.Any("error", err))
		}
	}
}
