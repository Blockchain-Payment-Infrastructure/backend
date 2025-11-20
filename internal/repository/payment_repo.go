package repository

import (
	"backend/internal/database"
	"backend/internal/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrorPaymentNotFound      = errors.New("payment not found")
	ErrorDuplicateTransaction = errors.New("transaction hash already exists")
	ErrorInvalidPaymentStatus = errors.New("invalid payment status")
)

// CreatePayment creates a new payment record in the database
func CreatePayment(ctx context.Context, payment *model.Payment) error {
	db := database.New("")

	query := `
		INSERT INTO payments (
			id, user_id, from_address, to_address, amount, currency,
			transaction_hash, block_number, gas_used, gas_price,
			status, description, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)`

	_, err := db.ExecContext(ctx, query,
		payment.ID,
		payment.UserID,
		payment.FromAddress,
		payment.ToAddress,
		payment.Amount,
		payment.Currency,
		payment.TransactionHash,
		payment.BlockNumber,
		payment.GasUsed,
		payment.GasPrice,
		payment.Status,
		payment.Description,
		payment.CreatedAt,
		payment.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique constraint violation
				if pgErr.ConstraintName == "payments_transaction_hash_key" {
					return ErrorDuplicateTransaction
				}
			}
		}
		slog.Error("Failed to create payment", slog.Any("error", err))
		return fmt.Errorf("%w: %v", ErrorDatabase, err)
	}

	return nil
}

// GetPaymentByID retrieves a payment by its ID
func GetPaymentByID(ctx context.Context, paymentID uuid.UUID) (*model.Payment, error) {
	db := database.New("")

	query := `
		SELECT id, user_id, from_address, to_address, amount, currency,
			   transaction_hash, block_number, gas_used, gas_price,
			   status, description, created_at, updated_at, confirmed_at
		FROM payments
		WHERE id = $1`

	payment := &model.Payment{}
	err := db.QueryRowContext(ctx, query, paymentID).Scan(
		&payment.ID,
		&payment.UserID,
		&payment.FromAddress,
		&payment.ToAddress,
		&payment.Amount,
		&payment.Currency,
		&payment.TransactionHash,
		&payment.BlockNumber,
		&payment.GasUsed,
		&payment.GasPrice,
		&payment.Status,
		&payment.Description,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.ConfirmedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrorPaymentNotFound
		}
		slog.Error("Failed to get payment by ID", slog.Any("error", err), slog.Any("paymentID", paymentID))
		return nil, fmt.Errorf("%w: %v", ErrorDatabase, err)
	}

	return payment, nil
}

// GetPaymentByTransactionHash retrieves a payment by its transaction hash
func GetPaymentByTransactionHash(ctx context.Context, txHash string) (*model.Payment, error) {
	db := database.New("")

	query := `
		SELECT id, user_id, from_address, to_address, amount, currency,
			   transaction_hash, block_number, gas_used, gas_price,
			   status, description, created_at, updated_at, confirmed_at
		FROM payments
		WHERE transaction_hash = $1`

	payment := &model.Payment{}
	err := db.QueryRowContext(ctx, query, txHash).Scan(
		&payment.ID,
		&payment.UserID,
		&payment.FromAddress,
		&payment.ToAddress,
		&payment.Amount,
		&payment.Currency,
		&payment.TransactionHash,
		&payment.BlockNumber,
		&payment.GasUsed,
		&payment.GasPrice,
		&payment.Status,
		&payment.Description,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.ConfirmedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrorPaymentNotFound
		}
		slog.Error("Failed to get payment by transaction hash", slog.Any("error", err), slog.String("txHash", txHash))
		return nil, fmt.Errorf("%w: %v", ErrorDatabase, err)
	}

	return payment, nil
}

// GetPaymentsByUserID retrieves payments for a specific user with pagination and filtering
func GetPaymentsByUserID(ctx context.Context, userID uuid.UUID, query *model.PaymentQuery) (*model.PaymentListResponse, error) {
	db := database.New("")

	// Build WHERE clause
	whereClause := "WHERE user_id = $1"
	args := []interface{}{userID}
	argCount := 1

	if query.Status != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *query.Status)
	}

	if query.Currency != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND currency = $%d", argCount)
		args = append(args, *query.Currency)
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM payments %s", whereClause)
	var totalCount int64
	err := db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		slog.Error("Failed to count payments", slog.Any("error", err))
		return nil, fmt.Errorf("%w: %v", ErrorDatabase, err)
	}

	// Calculate pagination
	offset := (query.Page - 1) * query.PageSize
	totalPages := int(math.Ceil(float64(totalCount) / float64(query.PageSize)))

	// Get paginated results
	selectQuery := fmt.Sprintf(`
		SELECT id, user_id, from_address, to_address, amount, currency,
			   transaction_hash, block_number, gas_used, gas_price,
			   status, description, created_at, updated_at, confirmed_at
		FROM payments %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argCount+1, argCount+2)

	args = append(args, query.PageSize, offset)

	rows, err := db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		slog.Error("Failed to query payments", slog.Any("error", err))
		return nil, fmt.Errorf("%w: %v", ErrorDatabase, err)
	}
	defer rows.Close()

	var payments []model.PaymentResponse
	for rows.Next() {
		payment := model.Payment{}
		err := rows.Scan(
			&payment.ID,
			&payment.UserID,
			&payment.FromAddress,
			&payment.ToAddress,
			&payment.Amount,
			&payment.Currency,
			&payment.TransactionHash,
			&payment.BlockNumber,
			&payment.GasUsed,
			&payment.GasPrice,
			&payment.Status,
			&payment.Description,
			&payment.CreatedAt,
			&payment.UpdatedAt,
			&payment.ConfirmedAt,
		)
		if err != nil {
			slog.Error("Failed to scan payment row", slog.Any("error", err))
			return nil, fmt.Errorf("%w: %v", ErrorDatabase, err)
		}
		payments = append(payments, payment.ToResponse())
	}

	if err = rows.Err(); err != nil {
		slog.Error("Row iteration error", slog.Any("error", err))
		return nil, fmt.Errorf("%w: %v", ErrorDatabase, err)
	}

	return &model.PaymentListResponse{
		Payments:   payments,
		TotalCount: totalCount,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdatePaymentStatus updates the status of a payment
func UpdatePaymentStatus(ctx context.Context, paymentID uuid.UUID, status model.PaymentStatus, blockNumber *int64, gasUsed *int64, gasPrice *string) error {
	if !status.IsValid() {
		return ErrorInvalidPaymentStatus
	}

	db := database.New("")

	var confirmedAt *time.Time
	if status == model.PaymentStatusConfirmed {
		now := time.Now()
		confirmedAt = &now
	}

	query := `
		UPDATE payments
		SET status = $1, block_number = $2, gas_used = $3, gas_price = $4, confirmed_at = $5, updated_at = NOW()
		WHERE id = $6`

	result, err := db.ExecContext(ctx, query, status, blockNumber, gasUsed, gasPrice, confirmedAt, paymentID)
	if err != nil {
		slog.Error("Failed to update payment status", slog.Any("error", err), slog.Any("paymentID", paymentID))
		return fmt.Errorf("%w: %v", ErrorDatabase, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrorDatabase, err)
	}

	if rowsAffected == 0 {
		return ErrorPaymentNotFound
	}

	return nil
}

// DeletePayment deletes a payment (admin only operation)
func DeletePayment(ctx context.Context, paymentID uuid.UUID) error {
	db := database.New("")

	query := "DELETE FROM payments WHERE id = $1"
	result, err := db.ExecContext(ctx, query, paymentID)
	if err != nil {
		slog.Error("Failed to delete payment", slog.Any("error", err), slog.Any("paymentID", paymentID))
		return fmt.Errorf("%w: %v", ErrorDatabase, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrorDatabase, err)
	}

	if rowsAffected == 0 {
		return ErrorPaymentNotFound
	}

	return nil
}

// GetPendingPayments retrieves all pending payments (for background processing)
func GetPendingPayments(ctx context.Context) ([]*model.Payment, error) {
	db := database.New("")

	query := `
		SELECT id, user_id, from_address, to_address, amount, currency,
			   transaction_hash, block_number, gas_used, gas_price,
			   status, description, created_at, updated_at, confirmed_at
		FROM payments
		WHERE status = $1
		ORDER BY created_at ASC`

	rows, err := db.QueryContext(ctx, query, model.PaymentStatusPending)
	if err != nil {
		slog.Error("Failed to query pending payments", slog.Any("error", err))
		return nil, fmt.Errorf("%w: %v", ErrorDatabase, err)
	}
	defer rows.Close()

	var payments []*model.Payment
	for rows.Next() {
		payment := &model.Payment{}
		err := rows.Scan(
			&payment.ID,
			&payment.UserID,
			&payment.FromAddress,
			&payment.ToAddress,
			&payment.Amount,
			&payment.Currency,
			&payment.TransactionHash,
			&payment.BlockNumber,
			&payment.GasUsed,
			&payment.GasPrice,
			&payment.Status,
			&payment.Description,
			&payment.CreatedAt,
			&payment.UpdatedAt,
			&payment.ConfirmedAt,
		)
		if err != nil {
			slog.Error("Failed to scan pending payment row", slog.Any("error", err))
			return nil, fmt.Errorf("%w: %v", ErrorDatabase, err)
		}
		payments = append(payments, payment)
	}

	if err = rows.Err(); err != nil {
		slog.Error("Row iteration error for pending payments", slog.Any("error", err))
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return payments, nil
}

// GetUserWalletAddresses retrieves all wallet addresses for a user
func GetUserWalletAddresses(ctx context.Context, userID string) ([]string, error) {
	db := database.New("")

	// First get user's phone number
	var phoneNumber string
	userQuery := "SELECT phone_number FROM users WHERE id = $1"
	err := db.QueryRowContext(ctx, userQuery, userID).Scan(&phoneNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrorUserNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrorDatabase, err)
	}

	// Then get wallet addresses for that phone number
	walletQuery := "SELECT address FROM wallet_address_phone WHERE phone_number = $1"
	rows, err := db.QueryContext(ctx, walletQuery, phoneNumber)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrorDatabase, err)
	}
	defer rows.Close()

	var addresses []string
	for rows.Next() {
		var address string
			if err := rows.Scan(&address); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrorDatabase, err)
		}
		addresses = append(addresses, address)
	}

	return addresses, nil
}
