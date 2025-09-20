package repository

import (
	"backend/internal/database"
	"backend/internal/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
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
		errors.As(err, &pgErr)
		slog.Error("postgres error:",
			slog.String("code", pgErr.Code),
			slog.Any("err", pgErr))

		return nil, err
	}

	tempAddress := model.WalletAddress{}
	for rows.Next() {
		if err := rows.Scan(&tempAddress.Address); err != nil {
			slog.Error("postgres error:",
				slog.Any("err", err))

			return nil, err
		}

		addresses = append(addresses, tempAddress)
	}

	return addresses, nil
}
func GetPhoneNumberByUserID(userID int) (string, error) {
	var phoneNumber string

	// We assume 'db' is your global or passed-in database connection object.
	// Replace with your actual database query logic.
	db := database.New("")
	query := "SELECT phone_number FROM users WHERE id = ?"
	err := db.QueryRowContext(context.Background(), query, userID).Scan(&phoneNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("user with ID %d not found", userID)
		}
		return "", fmt.Errorf("database query error: %w", err)
	}
	return phoneNumber, nil
}
func InsertWalletAddressPhone(walletAddress, phoneNumber string) error {
	db := database.New("")
	query := `
		INSERT INTO wallet_address_phone (wallet_address, phone_number) 
		VALUES (?, ?)
	`
	_, err := db.ExecContext(context.Background(), query, walletAddress, phoneNumber)
	if err != nil {

		if strings.Contains(err.Error(), "Duplicate entry") {
			return fmt.Errorf("wallet address already exists")
		}
		return fmt.Errorf("database insert error: %w", err)
	}
	return nil
}
