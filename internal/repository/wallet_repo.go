package repository

import (
	"backend/internal/database"
	"backend/internal/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

func GetWalletAddressesFromPhone(ctx context.Context, phone string) ([]model.WalletAddress, error) {
	var addresses []model.WalletAddress

	db := database.New("")
	query := `SELECT address FROM wallet_address_phone WHERE phone_number = $1;`
	rows, err := db.QueryContext(ctx, query, phone)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return nil, fmt.Errorf("%w: %v", ErrorDatabase, pgErr)
		}
		return nil, fmt.Errorf("%w: %v", ErrorDatabase, err)
	}

	tempAddress := model.WalletAddress{}
	for rows.Next() {
		if err := rows.Scan(&tempAddress.Address); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrorDatabase, err)
		}

		addresses = append(addresses, tempAddress)
	}

	return addresses, nil
}

func GetPhoneNumberByUserID(ctx context.Context, userID string) (string, error) {
	db := database.New("")

	var phoneNumber string
	query := "SELECT phone_number FROM users WHERE id = $1;"
	err := db.QueryRowContext(ctx, query, userID).Scan(&phoneNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrorUserNotFound
		}

		return "", fmt.Errorf("%w: %v", ErrorDatabase, err)
	}

	return phoneNumber, nil
}

func InsertWalletAddressPhone(ctx context.Context, walletAddress, phoneNumber string) error {
	db := database.New("")
	query := `
		INSERT INTO wallet_address_phone (address, phone_number)
		VALUES ($1, $2)
	`
	_, err := db.ExecContext(ctx, query, walletAddress, phoneNumber)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// PostgreSQL unique constraint violation
			if pgErr.Code == "23505" {
				return ErrorWalletAddressAlreadyExists
			}
		}
		// Also check for MySQL duplicate entry or generic duplicate error
		if strings.Contains(err.Error(), "Duplicate entry") || 
		   strings.Contains(err.Error(), "duplicate key value") ||
		   strings.Contains(err.Error(), "UNIQUE constraint") {
			return ErrorWalletAddressAlreadyExists
		}

		return fmt.Errorf("%w: %v", ErrorDatabase, err)
	}

	return nil
}
