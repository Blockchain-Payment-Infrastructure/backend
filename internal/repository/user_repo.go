package repository

import (
	"backend/internal/database"
	"backend/internal/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func CreateUser(ctx context.Context, user model.UserSignUp) error {
	db := database.New("")

	query := `
		INSERT INTO users (id, username, email, phone_number, password_hash, created_at)
		VALUES ($1, $2, $3, $4, $5, now())
	`

	_, err := db.ExecContext(ctx, query, uuid.New(), user.Username, user.Email, user.PhoneNumber, user.Password)
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
				case "users_phone_number_key":
					return ErrorPhoneNumberExists
				default:
					slog.Warn("Unhandled unique violation:",
						slog.String("constraint", pgErr.ConstraintName))
					return fmt.Errorf("%w: %s", ErrorUnhandledUniqueConstraint, pgErr.ConstraintName)
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

func FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}

	rawDB := database.New("")

	query := "SELECT id, email, username, phone_number, password_hash FROM users WHERE email = $1"
	row := rawDB.QueryRowContext(ctx, query, email)
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.PhoneNumber, &user.HashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrorUserNotFound
		}
		slog.Error("Database query error in findUserByEmail", slog.Any("error", err))
		return nil, fmt.Errorf("%w: %v", ErrorDatabase, err)
	}

	return user, nil
}

// FindUserByID finds a user by their numeric ID.
func FindUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	user := &model.User{}

	rawDB := database.New("")

	query := "SELECT id, email, username, phone_number, password_hash FROM users WHERE id = $1"
	row := rawDB.QueryRowContext(ctx, query, userID)
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.PhoneNumber, &user.HashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrorUserNotFound
		}
		slog.Error("Database query error in FindUserByID", slog.Any("error", err))
<<<<<<< HEAD
		return nil, fmt.Errorf("database query error: %w", err)
=======
		return nil, fmt.Errorf("%w: %v", ErrorDatabase, err)
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
	}

	return user, nil
}
func UpdateUserPassword(ctx context.Context, userID uuid.UUID, newHashedPassword string) error {
	db := database.New("")
	if db == nil {
<<<<<<< HEAD
		return fmt.Errorf("database service not set in repository")
=======
		return ErrorDatabaseServiceNotSet
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
	}

	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	result, err := database.New("").ExecContext(ctx, query, newHashedPassword, userID)
	if err != nil {
		slog.Error("Repository: Failed to update user password", slog.Any("error", err), slog.String("userID", userID.String()))
<<<<<<< HEAD
		return fmt.Errorf("failed to update password: %w", err)
=======
		return fmt.Errorf("%w: %v", ErrorUpdatePasswordFailed, err)
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
<<<<<<< HEAD
		return fmt.Errorf("user with ID %s not found or password was not changed", userID.String())
=======
		return ErrorUserNotModified
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
	}
	return nil
}

func UpdateUserEmail(ctx context.Context, userID uuid.UUID, newEmail string) error {
	db := database.New("")
	if db == nil {
<<<<<<< HEAD
		return fmt.Errorf("database service not set in repository")
=======
		return ErrorDatabaseServiceNotSet
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
	}

	query := `UPDATE users SET email = $1, updated_at = NOW() WHERE id = $2`
	result, err := db.ExecContext(ctx, query, newEmail, userID)
	if err != nil {
		slog.Error("Repository: Failed to update user email", slog.Any("error", err), slog.String("userID", userID.String()))
		// Check for unique constraint violation on email
		// if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" { // Example for pgx, adjust if needed
		// 	return fmt.Errorf("email already in use")
		// }
<<<<<<< HEAD
		return fmt.Errorf("failed to update email: %w", err)
=======
		return fmt.Errorf("%w: %v", ErrorUpdateEmailFailed, err)
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
<<<<<<< HEAD
		return fmt.Errorf("user with ID %s not found or email was not changed", userID.String())
=======
		return ErrorUserNotModified
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
	}
	return nil
}
func DeleteUser(ctx context.Context, userID uuid.UUID) error {
	db := database.New("")
	if db == nil {
<<<<<<< HEAD
		return fmt.Errorf("database service not set in repository")
=======
		return ErrorDatabaseServiceNotSet
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
	}

	// Start a transaction for deleting user and related data
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("Repository: Failed to begin transaction for user deletion", slog.Any("error", err))
<<<<<<< HEAD
		return fmt.Errorf("failed to begin transaction: %w", err)
=======
		return fmt.Errorf("%w: %v", ErrorDatabase, err)
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
	}
	defer tx.Rollback() // Rollback on error if commit doesn't happen

	// Example: Delete wallet addresses associated with the user first (if they exist)
	// You need to adjust this based on your actual schema and foreign key constraints
	// queryDeleteWallets := `DELETE FROM user_wallet_addresses WHERE user_id = $1`
	// _, err = tx.ExecContext(ctx, queryDeleteWallets, userID)
	// if err != nil {
	// 	slog.Error("Repository: Failed to delete user wallet addresses", slog.Any("error", err), slog.Uint64("userID", uint64(userID)))
	// 	return fmt.Errorf("failed to delete user wallet addresses: %w", err)
	// }

	// Then, delete the user itself
	queryDeleteUser := `DELETE FROM users WHERE id = $1`
	result, err := tx.ExecContext(ctx, queryDeleteUser, userID)
	if err != nil {
		slog.Error("Repository: Failed to delete user", slog.Any("error", err), slog.String("userID", userID.String()))
<<<<<<< HEAD
		return fmt.Errorf("failed to delete user: %w", err)
=======
		return fmt.Errorf("%w: %v", ErrorDatabase, err)
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
<<<<<<< HEAD
		return fmt.Errorf("user with ID %s not found or account was not deleted", userID.String())
=======
		return ErrorUserNotModified
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
	}

	return tx.Commit() // Commit the transaction
}
