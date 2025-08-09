package repository

import (
	"backend/internal/database"
	"backend/internal/model"
	"context"
	"log/slog"

	"github.com/google/uuid"
)

func CreateUser(ctx context.Context, user model.UserSignUp) error {
	db := database.New()
	query := `INSERT INTO users (id, username, email, password_hash, created_at) VALUES ($1, $2, $3, $4, now())`
	if _, err := db.ExecContext(ctx, query, uuid.New(), user.Username, user.Email, user.Password); err != nil {
		slog.Error("create user: ", slog.Any("error", err))
		return ErrorDatabase
	}

	return nil
}
