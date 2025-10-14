package model

import (
	"time"

	"github.com/google/uuid"
)

// PaymentStatus represents the status of a payment transaction
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusConfirmed PaymentStatus = "confirmed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

// Payment represents a blockchain payment transaction
type Payment struct {
	ID              uuid.UUID     `json:"id" db:"id"`
	UserID          uuid.UUID     `json:"user_id" db:"user_id"`
	FromAddress     string        `json:"from_address" db:"from_address"`
	ToAddress       string        `json:"to_address" db:"to_address"`
	Amount          string        `json:"amount" db:"amount"` // Using string to preserve precision
	Currency        string        `json:"currency" db:"currency"`
	TransactionHash string        `json:"transaction_hash" db:"transaction_hash"`
	BlockNumber     *int64        `json:"block_number,omitempty" db:"block_number"`
	GasUsed         *int64        `json:"gas_used,omitempty" db:"gas_used"`
	GasPrice        *string       `json:"gas_price,omitempty" db:"gas_price"`
	Status          PaymentStatus `json:"status" db:"status"`
	Description     *string       `json:"description,omitempty" db:"description"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at" db:"updated_at"`
	ConfirmedAt     *time.Time    `json:"confirmed_at,omitempty" db:"confirmed_at"`
}

// CreatePaymentRequest represents the request to create a new payment
type CreatePaymentRequest struct {
	ToAddress       string  `json:"to_address" binding:"required,len=42"` // Ethereum address length
	Amount          string  `json:"amount" binding:"required"`
	Currency        string  `json:"currency,omitempty"`
	TransactionHash string  `json:"transaction_hash" binding:"required,len=66"` // Transaction hash length
	Description     *string `json:"description,omitempty"`
}

// PaymentResponse represents the response after creating/retrieving a payment
type PaymentResponse struct {
	ID              uuid.UUID     `json:"id"`
	FromAddress     string        `json:"from_address"`
	ToAddress       string        `json:"to_address"`
	Amount          string        `json:"amount"`
	Currency        string        `json:"currency"`
	TransactionHash string        `json:"transaction_hash"`
	BlockNumber     *int64        `json:"block_number,omitempty"`
	Status          PaymentStatus `json:"status"`
	Description     *string       `json:"description,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	ConfirmedAt     *time.Time    `json:"confirmed_at,omitempty"`
}

// PaymentListResponse represents a paginated list of payments
type PaymentListResponse struct {
	Payments   []PaymentResponse `json:"payments"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// PaymentQuery represents query parameters for filtering payments
type PaymentQuery struct {
	Status   *PaymentStatus `form:"status"`
	Currency *string        `form:"currency"`
	Page     int            `form:"page,default=1"`
	PageSize int            `form:"page_size,default=20"`
}

// TransactionDetails represents detailed information about a blockchain transaction
type TransactionDetails struct {
	Hash        string `json:"hash"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	Gas         uint64 `json:"gas"`
	GasPrice    string `json:"gas_price"`
	BlockNumber *int64 `json:"block_number"`
	Status      uint64 `json:"status"` // 1 for success, 0 for failure
}

// ToResponse converts a Payment model to PaymentResponse
func (p *Payment) ToResponse() PaymentResponse {
	return PaymentResponse{
		ID:              p.ID,
		FromAddress:     p.FromAddress,
		ToAddress:       p.ToAddress,
		Amount:          p.Amount,
		Currency:        p.Currency,
		TransactionHash: p.TransactionHash,
		BlockNumber:     p.BlockNumber,
		Status:          p.Status,
		Description:     p.Description,
		CreatedAt:       p.CreatedAt,
		ConfirmedAt:     p.ConfirmedAt,
	}
}

// IsValidStatus checks if the payment status is valid
func (ps PaymentStatus) IsValid() bool {
	switch ps {
	case PaymentStatusPending, PaymentStatusConfirmed, PaymentStatusFailed, PaymentStatusCancelled:
		return true
	default:
		return false
	}
}
