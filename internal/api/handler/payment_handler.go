package handler

import (
	"errors"
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/service"
	"backend/internal/repository"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreatePaymentHandler godoc
//
//	@Summary		Create Payment
//	@Description	Creates a new payment record after verifying the transaction on blockchain
//	@Tags			payments
//	@Accept			json
//	@Produce		json
//	@Param			paymentRequest	body		model.CreatePaymentRequest	true	"Payment creation details"
//	@Success		201				{object}	model.PaymentResponse		"Payment created successfully"
//	@Failure		400				{string}	string						"Validation error or bad request"
//	@Failure		401				{string}	string						"Unauthorized"
//	@Failure		500				{string}	string						"Internal server error"
//	@Router			/payments [post]
//	@Security		BearerAuth
func CreatePaymentHandler(c *gin.Context) {
	userID, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		JSONError(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		JSONError(c, http.StatusInternalServerError, "Invalid user ID in context", nil)
		return
	}

	var req model.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, "Invalid request payload: "+err.Error(), err)
		return
	}

	payment, err := service.CreatePayment(c.Request.Context(), userIDStr, &req)
	if err != nil {
		slog.Error("Failed to create payment", slog.Any("error", err), slog.String("userID", userIDStr))
		// Map known repository errors
		if errors.Is(err, repository.ErrorDuplicateTransaction) {
			JSONError(c, http.StatusConflict, err.Error(), err)
			return
		}
		JSONError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	JSONSuccess(c, http.StatusCreated, gin.H{
		"message": "Payment created successfully",
		"payment": payment,
	})
}

// GetPaymentHandler godoc
//
//	@Summary		Get Payment
//	@Description	Retrieves a specific payment by ID
//	@Tags			payments
//	@Produce		json
//	@Param			id	path		string					true	"Payment ID"
//	@Success		200	{object}	model.PaymentResponse	"Payment details"
//	@Failure		400	{string}	string					"Invalid payment ID"
//	@Failure		401	{string}	string					"Unauthorized"
//	@Failure		404	{string}	string					"Payment not found"
//	@Failure		500	{string}	string					"Internal server error"
//	@Router			/payments/{id} [get]
//	@Security		BearerAuth
func GetPaymentHandler(c *gin.Context) {
	userID, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		JSONError(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		JSONError(c, http.StatusInternalServerError, "Invalid user ID in context", nil)
		return
	}

	paymentID := c.Param("id")
	if _, err := uuid.Parse(paymentID); err != nil {
		JSONError(c, http.StatusBadRequest, "Invalid payment ID format", err)
		return
	}

	payment, err := service.GetPayment(c.Request.Context(), userIDStr, paymentID)
	if err != nil {
		if errors.Is(err, repository.ErrorPaymentNotFound) {
			JSONError(c, http.StatusNotFound, "Payment not found", err)
			return
		}

		slog.Error("Failed to get payment", slog.Any("error", err), slog.String("paymentID", paymentID))
		JSONError(c, http.StatusInternalServerError, "Failed to retrieve payment", err)
		return
	}

	JSONSuccess(c, http.StatusOK, payment)
}

// GetUserPaymentsHandler godoc
//
//	@Summary		Get User Payments
//	@Description	Retrieves paginated list of payments for the authenticated user
//	@Tags			payments
//	@Produce		json
//	@Param			status		query		string						false	"Filter by payment status"
//	@Param			currency	query		string						false	"Filter by currency"
//	@Param			page		query		int							false	"Page number (default: 1)"
//	@Param			page_size	query		int							false	"Page size (default: 20, max: 100)"
//	@Success		200			{object}	model.PaymentListResponse	"List of payments"
//	@Failure		400			{string}	string						"Invalid query parameters"
//	@Failure		401			{string}	string						"Unauthorized"
//	@Failure		500			{string}	string						"Internal server error"
//	@Router			/payments [get]
//	@Security		BearerAuth
func GetUserPaymentsHandler(c *gin.Context) {
	userID, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		JSONError(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		JSONError(c, http.StatusInternalServerError, "Invalid user ID in context", nil)
		return
	}

	var query model.PaymentQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		JSONError(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error(), err)
		return
	}

	payments, err := service.GetUserPayments(c.Request.Context(), userIDStr, &query)
	if err != nil {
		slog.Error("Failed to get user payments", slog.Any("error", err), slog.String("userID", userIDStr))
		JSONError(c, http.StatusInternalServerError, "Failed to retrieve payments", err)
		return
	}
	JSONSuccess(c, http.StatusOK, payments)
}

// GetPaymentByTransactionHashHandler godoc
//
//	@Summary		Get Payment by Transaction Hash
//	@Description	Retrieves a payment by its blockchain transaction hash
//	@Tags			payments
//	@Produce		json
//	@Param			hash	path		string					true	"Transaction Hash"
//	@Success		200		{object}	model.PaymentResponse	"Payment details"
//	@Failure		400		{string}	string					"Invalid transaction hash"
//	@Failure		401		{string}	string					"Unauthorized"
//	@Failure		404		{string}	string					"Payment not found"
//	@Failure		500		{string}	string					"Internal server error"
//	@Router			/payments/tx/{hash} [get]
//	@Security		BearerAuth
func GetPaymentByTransactionHashHandler(c *gin.Context) {
	userID, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		JSONError(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		JSONError(c, http.StatusInternalServerError, "Invalid user ID in context", nil)
		return
	}

	txHash := c.Param("hash")
	if txHash == "" {
		JSONError(c, http.StatusBadRequest, "Transaction hash is required", nil)
		return
	}

	// Basic validation for transaction hash format
	if len(txHash) != 66 || txHash[:2] != "0x" {
		JSONError(c, http.StatusBadRequest, "Invalid transaction hash format", nil)
		return
	}

	payment, err := service.GetPaymentByTransactionHash(c.Request.Context(), userIDStr, txHash)
	if err != nil {
		if errors.Is(err, repository.ErrorPaymentNotFound) {
			JSONError(c, http.StatusNotFound, "Payment not found", err)
			return
		}
		slog.Error("Failed to get payment by transaction hash", slog.Any("error", err), slog.String("txHash", txHash))
		JSONError(c, http.StatusInternalServerError, "Failed to retrieve payment", err)
		return
	}
	JSONSuccess(c, http.StatusOK, payment)
}

// RefreshPaymentStatusHandler godoc
//
//	@Summary		Refresh Payment Status
//	@Description	Checks blockchain for payment status updates and refreshes the payment record
//	@Tags			payments
//	@Produce		json
//	@Param			id	path		string					true	"Payment ID"
//	@Success		200	{object}	model.PaymentResponse	"Updated payment details"
//	@Failure		400	{string}	string					"Invalid payment ID"
//	@Failure		401	{string}	string					"Unauthorized"
//	@Failure		404	{string}	string					"Payment not found"
//	@Failure		500	{string}	string					"Internal server error"
//	@Router			/payments/{id}/refresh [post]
//	@Security		BearerAuth
func RefreshPaymentStatusHandler(c *gin.Context) {
	userID, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		JSONError(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		JSONError(c, http.StatusInternalServerError, "Invalid user ID in context", nil)
		return
	}

	paymentID := c.Param("id")
	if paymentID == "" {
		JSONError(c, http.StatusBadRequest, "Payment ID is required", nil)
		return
	}

	payment, err := service.RefreshPaymentStatus(c.Request.Context(), userIDStr, paymentID)
	if err != nil {
		if errors.Is(err, repository.ErrorPaymentNotFound) {
			JSONError(c, http.StatusNotFound, "Payment not found", err)
			return
		}
		slog.Error("Failed to refresh payment status", slog.Any("error", err), slog.String("paymentID", paymentID))
		JSONError(c, http.StatusInternalServerError, "Failed to refresh payment status", err)
		return
	}

	JSONSuccess(c, http.StatusOK, gin.H{
		"message": "Payment status refreshed",
		"payment": payment,
	})
}

// GetPaymentStatsHandler godoc
//
//	@Summary		Get Payment Statistics
//	@Description	Retrieves payment statistics for the authenticated user
//	@Tags			payments
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Payment statistics"
//	@Failure		401	{string}	string					"Unauthorized"
//	@Failure		500	{string}	string					"Internal server error"
//	@Router			/payments/stats [get]
//	@Security		BearerAuth
func GetPaymentStatsHandler(c *gin.Context) {
	userID, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		JSONError(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		JSONError(c, http.StatusInternalServerError, "Invalid user ID in context", nil)
		return
	}

	stats, err := service.GetPaymentStats(c.Request.Context(), userIDStr)
	if err != nil {
		slog.Error("Failed to get payment stats", slog.Any("error", err), slog.String("userID", userIDStr))
		JSONError(c, http.StatusInternalServerError, "Failed to retrieve payment statistics", err)
		return
	}
	JSONSuccess(c, http.StatusOK, stats)
}
