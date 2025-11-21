package handler

import (
	"backend/internal/model"
	"backend/internal/service"
	//"backend/internal/utils"
	//"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Helper to get UserID from Gin context
func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDVal, exists := c.Get("userID") // AuthMiddleware should set `userID` as uuid.UUID or string
	if !exists {
		slog.Error("Handler: UserID not found in context (AuthMiddleware didn't set it). Ensure the request includes 'Authorization: Bearer <access_token>' and that the token contains a 'user_id' claim.")
		return uuid.Nil, fmt.Errorf("user ID not found in context; ensure Authorization header 'Bearer <token>' is present")
	}

	// Accept both uuid.UUID and string representations for flexibility
	switch v := userIDVal.(type) {
	case uuid.UUID:
		return v, nil
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			slog.Error("Handler: failed to parse userID string", slog.Any("userIDVal", v), slog.Any("error", err))
			return uuid.Nil, fmt.Errorf("invalid user ID in context")
		}
		return id, nil
	default:
		slog.Error("Handler: UserID in context is not a UUID or string", slog.Any("type", fmt.Sprintf("%T", userIDVal)))
		return uuid.Nil, fmt.Errorf("invalid user ID type in context")
	}
}

// ChangePasswordHandler handles requests to change a user's password.
func ChangePasswordHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req model.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.ChangePasswordService(c.Request.Context(), userID, req); err != nil {
		// Differentiate error responses for better UX
		if err.Error() == "invalid old password" || err.Error() == "user not found or invalid credentials" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		} else if err.Error() == "password must be at least 8 characters long" { // Or other specific validation errors
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			slog.Error("Handler: ChangePassword failed unexpectedly", slog.String("userID", userID.String()), slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// UpdateEmailHandler handles requests to update a user's email.
func UpdateEmailHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req model.UpdateEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.UpdateEmailService(c.Request.Context(), userID, req); err != nil {
		if err.Error() == "invalid password" || err.Error() == "user not found or invalid credentials" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		} else if err.Error() == "email already in use by another account" || err.Error() == "invalid email format" { // Or other specific validation errors
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()}) // Use 409 Conflict for uniqueness
		} else {
			slog.Error("Handler: UpdateEmail failed unexpectedly", slog.String("userID", userID.String()), slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update email"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email updated successfully"})
}

// DeleteAccountHandler handles requests to delete a user's account.
func DeleteAccountHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req model.DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.DeleteAccountService(c.Request.Context(), userID, req); err != nil {
		if err.Error() == "invalid password" || err.Error() == "user not found or invalid credentials" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		} else {
			slog.Error("Handler: DeleteAccount failed unexpectedly", slog.String("userID", userID.String()), slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}
