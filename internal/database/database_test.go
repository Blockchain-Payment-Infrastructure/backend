package database

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testDSN string

// mustStartPostgresContainer starts a Postgres container and returns teardown func + DSN
func mustStartPostgresContainer() (func(context.Context, ...testcontainers.TerminateOption) error, string, error) {
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

func TestMain(m *testing.M) {
	teardown, dsn, err := mustStartPostgresContainer()
	if err != nil {
		slog.Error("could not start postgres container:", slog.Any("error", err))
		os.Exit(1)
	}

	testDSN = dsn

	code := m.Run()

	if teardown != nil {
		if err := teardown(context.Background()); err != nil {
			slog.Error("could not teardown postgres container:", slog.Any("error", err))
		}
	}

	os.Exit(code)
}

func TestNew(t *testing.T) {
	srv := New(testDSN)
	if srv == nil {
		t.Fatal("New() returned nil")
	}
}

func TestHealth(t *testing.T) {
	srv := New(testDSN)

	stats := srv.Health()

	if stats["status"] != "up" {
		t.Fatalf("expected status to be up, got %s", stats["status"])
	}

	if _, ok := stats["error"]; ok {
		t.Fatalf("expected error not to be present")
	}

	if stats["message"] != "It's healthy" {
		t.Fatalf("expected message to be 'It's healthy', got %s", stats["message"])
	}
}

func TestClose(t *testing.T) {
	srv := New(testDSN)

	if srv.Close() != nil {
		t.Fatalf("expected Close() to return nil")
	}
}
