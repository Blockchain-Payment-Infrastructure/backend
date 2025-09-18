package repository

import (
	"backend/internal/database"
	"backend/internal/model"
	"context"
	"errors"
	"log/slog"

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
