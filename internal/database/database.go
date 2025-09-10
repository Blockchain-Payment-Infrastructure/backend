// File: backend/internal/database/database.go (MODIFIED FOR POSTGRES & GETRAWDB)
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres" // PostgreSQL specific
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	_ "github.com/joho/godotenv/autoload"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error

	// ADDED: Method to get the raw *sql.DB connection
	GetRawDB() *sql.DB // <--- ADDED THIS

	// context-aware primitives
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row

	// transactions
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type service struct {
	db *sql.DB
}

var (
	// Removed global database connection details variables as they're now handled in New()
	dbInstance *service // Keep this for your singleton pattern
)

func Migrate(migratePath string) {
	slog.Info("Beginning Database Migration")
	// Ensure the DB instance is available for migrations
	if dbInstance == nil {
		slog.Error("Database service not initialized before migration. Call database.New() first.")
		os.Exit(1)
	}

	driver, err := postgres.WithInstance(dbInstance.db, &postgres.Config{})
	if err != nil {
		slog.Error("migration error:", slog.Any("error", err))
		os.Exit(1)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migratePath, // Use the provided path directly
		"postgres", driver,
	)
	if len(migratePath) == 0 { // Fallback if migratePath is empty
		m, err = migrate.NewWithDatabaseInstance(
			"file://db/migrations",
			"postgres", driver,
		)
	}

	if err != nil {
		slog.Error("migration error:", slog.Any("error", err))
		os.Exit(1)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		slog.Error("migration error:", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("Finished Database Migration")
}

// New connects to the database. It now takes a dbUrl, if empty it uses environment variables.
// It acts as a singleton.
func New(dbUrl string) Service {
	if dbInstance != nil && dbInstance.db != nil { // Ensure DB connection is actually open
		return dbInstance
	}

	// If dbUrl is empty, construct it from environment variables
	if dbUrl == "" {
		user := os.Getenv("DB_USERNAME")
		pass := os.Getenv("DB_PASSWORD")
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		dbName := os.Getenv("DB_DATABASE")

		dbUrl = fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s",
			user, pass, host, port, dbName)

		if os.Getenv("GIN_MODE") == gin.ReleaseMode {
			dbUrl += " sslmode=require"
		} else {
			dbUrl += " sslmode=disable" // Often needed for local dev with PG
		}
	}

	db, err := sql.Open("pgx", dbUrl) // Use "pgx" driver for PostgreSQL
	if err != nil {
		slog.Error("database connection error:", slog.Any("error", err))
		os.Exit(1)
	}

	// Ping the database to verify the connection is alive
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		slog.Error("database ping failed:", slog.Any("error", err))
		db.Close() // Close the connection if ping fails
		os.Exit(1)
	}
	slog.Info("Database connection successful!")


	dbInstance = &service{db: db}
	return dbInstance
}

// ADDED: GetRawDB method for the service struct
func (s *service) GetRawDB() *sql.DB {
	return s.db
}


func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		slog.Error("db down:", slog.Any("error", err))
		os.Exit(1)
	}

	stats["status"] = "up"
	stats["message"] = "It's healthy"

	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	if dbStats.OpenConnections > 40 {
		stats["message"] = "The database is experiencing heavy load."
	}
	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}
	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}
	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

func (s *service) Close() error {
	slog.Info("Disconnected from database")
	// Clear the singleton instance upon closing
	dbInstance = nil
	return s.db.Close()
}

func (s *service) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return s.db.ExecContext(ctx, query, args...)
}

func (s *service) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, query, args...)
}

func (s *service) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return s.db.QueryRowContext(ctx, query, args...)
}

func (s *service) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return s.db.BeginTx(ctx, opts)
}