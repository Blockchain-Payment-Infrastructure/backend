package repository

import (
	"backend/internal/database"
	"backend/internal/model"
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func CreateUser(ctx context.Context, user model.UserSignUp) error {
	db := database.New()
	query := `INSERT INTO users (id, username, email, password_hash, created_at) VALUES ($1, $2, $3, $4, now())`
	if _, err := db.ExecContext(ctx, query, uuid.New(), user.Username, user.Email, user.Password); err != nil {
		error, _ := err.(*pgconn.PgError)
		switch error.Code {
		case "23505":
			{
				slog.Info(error.Error())
				return ErrorUserExists
			}
		default:
			{
				slog.Error("c", slog.Any("error", err))
			}

		}
	}

	return nil
}
