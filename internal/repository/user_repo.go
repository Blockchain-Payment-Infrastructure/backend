package repository

import (
	"backend/internal/database"
	"backend/internal/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func CreateUser(ctx context.Context, user model.UserSignUp) error {
	db := database.New("")

	query := `
		INSERT INTO users (id, username, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4, now())
	`

	_, err := db.ExecContext(ctx, query, uuid.New(), user.Username, user.Email, user.Password)
	if err != nil {
		// check if it's a Postgres error
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique constraint violation
				switch pgErr.ConstraintName {
				case "users_email_key":
					return ErrorEmailExists
				case "users_username_key":
					return ErrorUsernameExists
				default:
					slog.Warn("Unhandled unique violation:",
						slog.String("constraint", pgErr.ConstraintName))
					return fmt.Errorf("unhandled unique constraint: %s", pgErr.ConstraintName)
				}
			default:
				slog.Error("postgres error:",
					slog.String("code", pgErr.Code),
					slog.Any("err", pgErr))
				return pgErr
			}
		}
		return err
	}

	return nil
}

func UserExists(ctx context.Context, email string) (bool, error) {
	var count int

	db := database.New("")
	err := db.QueryRowContext(ctx, "SELECT COUNT(1) FROM users WHERE email=$1", email).Scan(&count)
	return count > 0, err
}

func FindUserByEmail(email string) (*model.User, error) {
	user := &model.User{}

	rawDB := database.New("")
	if rawDB == nil {
		slog.Error("Raw database connection is nil in auth handler.")
		return nil, gin.Error{Err: nil, Type: gin.ErrorTypePrivate, Meta: "Raw database connection unavailable"}
	}

	query := "SELECT id, email, username, password_hash FROM users WHERE email = $1"
	row := rawDB.QueryRowContext(context.Background(), query, email)
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.HashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, gin.Error{Err: err, Type: gin.ErrorTypePrivate, Meta: "User not found"}
		}
		slog.Error("Database query error in findUserByEmail", slog.Any("error", err))
		return nil, gin.Error{Err: err, Type: gin.ErrorTypePrivate, Meta: "Database query error"}
	}

	return user, nil
}
