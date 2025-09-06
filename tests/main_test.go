package tests

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"
)

var (
	postgresContainer testcontainers.Container
	testDSN           string
)

func TestMain(m *testing.M) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
	slog.SetDefault(logger)

	teardown, dsn, err := MustStartPostgresContainer()
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
