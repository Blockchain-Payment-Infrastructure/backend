package service

import (
	"backend/internal/ethclient"
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CreatePayment creates a new payment after verifying the transaction on blockchain
func CreatePayment(ctx context.Context, userID string, req *model.CreatePaymentRequest) (*model.PaymentResponse, error) {
	// Create Ethereum client
	ethClient, err := ethclient.NewClient()
	if err != nil {
		slog.Error("Failed to create Ethereum client", slog.Any("error", err))
		return nil, fmt.Errorf("failed to connect to blockchain: %w", err)
	}
	defer ethClient.Close()

	// Get user's wallet addresses
	userWallets, err := repository.GetUserWalletAddresses(ctx, userID)
	if err != nil {
		slog.Error("Failed to get user wallet addresses", slog.Any("error", err), slog.String("userID", userID))
		return nil, fmt.Errorf("failed to get user wallet addresses: %w", err)
	}

	if len(userWallets) == 0 {
		return nil, fmt.Errorf("no wallet addresses found for user")
	}

	// Check if transaction already exists
	existingPayment, err := repository.GetPaymentByTransactionHash(ctx, req.TransactionHash)
	if err == nil {
		slog.Warn("Transaction hash already exists", slog.String("txHash", req.TransactionHash))
		response := existingPayment.ToResponse()
		return &response, nil
	}
	if err != repository.ErrorPaymentNotFound {
		slog.Error("Failed to check existing payment", slog.Any("error", err))
		return nil, fmt.Errorf("failed to check existing payment: %w", err)
	}

	// Verify transaction on blockchain
	txDetails, err := ethClient.VerifyTransaction(ctx, req.TransactionHash)
	if err != nil {
		slog.Error("Failed to verify transaction", slog.Any("error", err), slog.String("txHash", req.TransactionHash))
		return nil, fmt.Errorf("failed to verify transaction: %w", err)
	}

	// Find which of user's wallets matches the transaction from address
	var fromAddress string
	for _, wallet := range userWallets {
		if equalAddresses(wallet, txDetails.From) {
			fromAddress = wallet
			break
		}
	}

	if fromAddress == "" {
		return nil, fmt.Errorf("transaction is not from any of your connected wallets")
	}

	// Validate transaction matches the request
	err = ethClient.ValidateTransactionForPayment(ctx, txDetails, req, fromAddress)
	if err != nil {
		slog.Error("Transaction validation failed", slog.Any("error", err))
		return nil, fmt.Errorf("transaction validation failed: %w", err)
	}

	// Parse user ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	// Set default currency if not provided
	currency := req.Currency
	if currency == "" {
		currency = "ETH"
	}

	// Create payment record
	payment := &model.Payment{
		ID:              uuid.New(),
		UserID:          userUUID,
		FromAddress:     txDetails.From,
		ToAddress:       txDetails.To,
		Amount:          txDetails.Value,
		Currency:        currency,
		TransactionHash: txDetails.Hash,
		BlockNumber:     txDetails.BlockNumber,
		Status:          model.PaymentStatusPending,
		Description:     req.Description,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// If transaction is confirmed, update status and add confirmation details
	if txDetails.Status == 1 && txDetails.BlockNumber != nil {
		payment.Status = model.PaymentStatusConfirmed
		now := time.Now()
		payment.ConfirmedAt = &now

		// Add gas information if available
		if txDetails.Gas > 0 {
			gasUsed := int64(txDetails.Gas)
			payment.GasUsed = &gasUsed
		}

		if txDetails.GasPrice != "" {
			payment.GasPrice = &txDetails.GasPrice
		}
	}

	// Save payment to database
	err = repository.CreatePayment(ctx, payment)
	if err != nil {
		slog.Error("Failed to create payment record", slog.Any("error", err))
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	slog.Info("Payment created successfully",
		slog.String("paymentID", payment.ID.String()),
		slog.String("txHash", payment.TransactionHash),
		slog.String("status", string(payment.Status)))

	response := payment.ToResponse()
	return &response, nil
}

// GetPayment retrieves a payment by ID for a specific user
func GetPayment(ctx context.Context, userID string, paymentID string) (*model.PaymentResponse, error) {
	paymentUUID, err := uuid.Parse(paymentID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID format: %w", err)
	}

	payment, err := repository.GetPaymentByID(ctx, paymentUUID)
	if err != nil {
		return nil, err
	}

	// Verify user owns this payment
	if payment.UserID.String() != userID {
		return nil, fmt.Errorf("payment not found")
	}

	response := payment.ToResponse()
	return &response, nil
}

// GetUserPayments retrieves all payments for a user with pagination and filtering
func GetUserPayments(ctx context.Context, userID string, query *model.PaymentQuery) (*model.PaymentListResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	// Validate query parameters
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}

	if query.Status != nil && !query.Status.IsValid() {
		return nil, fmt.Errorf("invalid status filter")
	}

	return repository.GetPaymentsByUserID(ctx, userUUID, query)
}

// UpdatePaymentStatus updates the status of a payment (internal use)
func UpdatePaymentStatus(ctx context.Context, paymentID uuid.UUID, status model.PaymentStatus) error {
	if !status.IsValid() {
		return fmt.Errorf("invalid payment status")
	}

	// If updating to confirmed, try to get latest blockchain info
	var blockNumber *int64
	var gasUsed *int64
	var gasPrice *string

	if status == model.PaymentStatusConfirmed {
		payment, err := repository.GetPaymentByID(ctx, paymentID)
		if err == nil {
			ethClient, err := ethclient.NewClient()
			if err == nil {
				defer ethClient.Close()

				txDetails, err := ethClient.VerifyTransaction(ctx, payment.TransactionHash)
				if err == nil {
					blockNumber = txDetails.BlockNumber
					if txDetails.Gas > 0 {
						gas := int64(txDetails.Gas)
						gasUsed = &gas
					}
					if txDetails.GasPrice != "" {
						gasPrice = &txDetails.GasPrice
					}
				}
			}
		}
	}

	return repository.UpdatePaymentStatus(ctx, paymentID, status, blockNumber, gasUsed, gasPrice)
}

// GetPaymentByTransactionHash retrieves a payment by transaction hash
func GetPaymentByTransactionHash(ctx context.Context, userID string, txHash string) (*model.PaymentResponse, error) {
	payment, err := repository.GetPaymentByTransactionHash(ctx, txHash)
	if err != nil {
		return nil, err
	}

	// Verify user owns this payment
	if payment.UserID.String() != userID {
		return nil, fmt.Errorf("payment not found")
	}

	response := payment.ToResponse()
	return &response, nil
}

// RefreshPaymentStatus checks blockchain for payment status updates
func RefreshPaymentStatus(ctx context.Context, userID string, paymentID string) (*model.PaymentResponse, error) {
	paymentUUID, err := uuid.Parse(paymentID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID format: %w", err)
	}

	payment, err := repository.GetPaymentByID(ctx, paymentUUID)
	if err != nil {
		return nil, err
	}

	// Verify user owns this payment
	if payment.UserID.String() != userID {
		return nil, fmt.Errorf("payment not found")
	}

	// Only refresh pending payments
	if payment.Status != model.PaymentStatusPending {
		response := payment.ToResponse()
		return &response, nil
	}

	// Check blockchain status
	ethClient, err := ethclient.NewClient()
	if err != nil {
		slog.Error("Failed to create Ethereum client for refresh", slog.Any("error", err))
		return nil, fmt.Errorf("failed to connect to blockchain: %w", err)
	}
	defer ethClient.Close()

	txDetails, err := ethClient.VerifyTransaction(ctx, payment.TransactionHash)
	if err != nil {
		slog.Error("Failed to verify transaction for refresh", slog.Any("error", err))
		// Return current payment if we can't verify
		response := payment.ToResponse()
		return &response, nil
	}

	// Update status if transaction is now confirmed
	if txDetails.Status == 1 && txDetails.BlockNumber != nil {
		var gasUsed *int64
		var gasPrice *string

		if txDetails.Gas > 0 {
			gas := int64(txDetails.Gas)
			gasUsed = &gas
		}
		if txDetails.GasPrice != "" {
			gasPrice = &txDetails.GasPrice
		}

		err = repository.UpdatePaymentStatus(ctx, paymentUUID, model.PaymentStatusConfirmed,
			txDetails.BlockNumber, gasUsed, gasPrice)
		if err != nil {
			slog.Error("Failed to update payment status", slog.Any("error", err))
		} else {
			// Refetch updated payment
			payment, err = repository.GetPaymentByID(ctx, paymentUUID)
			if err != nil {
				slog.Error("Failed to refetch updated payment", slog.Any("error", err))
			}
		}
	}

	response := payment.ToResponse()
	return &response, nil
}

// GetPaymentStats returns payment statistics for a user
func GetPaymentStats(ctx context.Context, userID string) (map[string]interface{}, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	// Get all payments for stats calculation
	query := &model.PaymentQuery{
		Page:     1,
		PageSize: 1000, // Get a large number for stats
	}

	payments, err := repository.GetPaymentsByUserID(ctx, userUUID, query)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_payments": payments.TotalCount,
		"confirmed":      0,
		"pending":        0,
		"failed":         0,
		"total_amount":   "0",
	}

	// Calculate stats
	totalAmountWei := int64(0)
	for _, payment := range payments.Payments {
		switch payment.Status {
		case model.PaymentStatusConfirmed:
			stats["confirmed"] = stats["confirmed"].(int) + 1
		case model.PaymentStatusPending:
			stats["pending"] = stats["pending"].(int) + 1
		case model.PaymentStatusFailed:
			stats["failed"] = stats["failed"].(int) + 1
		}

		// Add to total amount (assuming ETH/Wei)
		if payment.Status == model.PaymentStatusConfirmed {
			if amount, err := strconv.ParseInt(payment.Amount, 10, 64); err == nil {
				totalAmountWei += amount
			}
		}
	}

	stats["total_amount"] = strconv.FormatInt(totalAmountWei, 10)
	return stats, nil
}

// Helper function to compare Ethereum addresses
func equalAddresses(addr1, addr2 string) bool {
	// Simple case-insensitive comparison
	// In production, you might want to use ethereum's common.HexToAddress for proper comparison
	return len(addr1) == len(addr2) &&
		len(addr1) == 42 &&
		addr1[0:2] == "0x" &&
		addr2[0:2] == "0x" &&
		strings.EqualFold(addr1[2:], addr2[2:])
}
