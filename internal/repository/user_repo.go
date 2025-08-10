package repository

import (
	"backend/internal/database"
	"backend/internal/model"
	"context"
	"errors"
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
				switch error.ConstraintName {
				case "users_email_key":
					return ErrorEmailExists
				case "users_username_key":
					return ErrorUsernameExists
				default:
					slog.Warn("Implement the following key violation:", slog.Any("violation", error.ConstraintName))
					return errors.New("implement other key violations")
				}
			}
		default:
			{
				slog.Error("create user:", slog.Any("error", err))
			}

		}
	}

	return nil
}
