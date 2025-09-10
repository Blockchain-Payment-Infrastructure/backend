// File: backend/internal/service/auth_service.go (MODIFIED for Argon2ID hashing)
package service

import (
	"context"
	"backend/internal/database"
	"backend/internal/model"
	"errors"
	"fmt"
	"log/slog"

	"github.com/alexedwards/argon2id" // Import the argon2id package
)

// dbService needs to be injected into this package too
var dbService database.Service

// SetDBService is used by the server to inject the database connection into this service package
func SetDBService(service database.Service) {
	dbService = service
}

// SignUpService handles the business logic for user registration.
// It hashes the password and saves the user to the database.
func SignUpService(ctx context.Context, userDetails model.UserSignUp) error {
	if dbService == nil {
		return errors.New("database service not set in auth service")
	}

	// 1. Hash the password using argon2id.CreateHash
	hashedPassword, err := argon2id.CreateHash(userDetails.Password, argon2id.DefaultParams)
	if err != nil {
		slog.Error("Failed to hash password", slog.Any("error", err))
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 2. Insert into database
	query := `INSERT INTO users (email, username, hashed_password) VALUES ($1, $2, $3)`
	_, err = dbService.ExecContext(ctx, query, userDetails.Email, userDetails.Username, hashedPassword)
	if err != nil {
		slog.Error("Failed to insert new user", slog.Any("error", err))
		// You might want more specific error handling here, e.g., checking for unique constraint violations
		return fmt.Errorf("failed to create user: %w", err)
	}
	slog.Info("User created successfully", slog.String("email", userDetails.Email), slog.String("username", userDetails.Username))
	return nil
}