package repository

import (
	"backend/internal/database"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	TokenHash string    `db:"token_hash"`
	ExpiresAt time.Time `db:"expires_at"`
}

// HashRefreshToken creates a SHA-256 hash of a token.
func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// StoreRefreshToken saves a new refresh token hash to the database.
func StoreRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	db := database.New("")
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := db.ExecContext(ctx, query, userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrorRefreshTokenStoreFailed, err)
	}
	return nil
}

// GetUserByRefreshToken retrieves a user ID by a refresh token hash.
// It also checks if the token is expired.
func GetUserByRefreshToken(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	db := database.New("")
	var token RefreshToken
	query := `
		SELECT user_id, expires_at FROM refresh_tokens
		WHERE token_hash = $1
	`
	err := db.QueryRowContext(ctx, query, tokenHash).Scan(&token.UserID, &token.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, ErrorRefreshTokenNotFound
		}
		return uuid.Nil, fmt.Errorf("%w: %v", ErrorRefreshTokenStoreFailed, err)
	}

	if time.Now().After(token.ExpiresAt) {
		return uuid.Nil, ErrorRefreshTokenExpired
	}

	return token.UserID, nil
}

// DeleteRefreshToken deletes a refresh token from the database.
func DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	db := database.New("")
	query := "DELETE FROM refresh_tokens WHERE token_hash = $1"
	_, err := db.ExecContext(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrorRefreshTokenDeleteFailed, err)
	}
	return nil
}
