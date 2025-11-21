package handler

import (
	"backend/internal/model"
	"backend/internal/service"
	"backend/internal/utils"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		slog.Error("UserID not found in context")
		return uuid.Nil, fmt.Errorf("user ID not found in context; ensure Authorization header 'Bearer <token>' is present")
	}

	userID := userIDVal.(string)
	id, err := uuid.Parse(userID)
	if err != nil {
		slog.Error("Failed to parse userID string", slog.Any("userIDVal", userID), slog.Any("error", err))
		return uuid.Nil, fmt.Errorf("invalid user ID in context")
	}

	return id, nil
}

func ChangePasswordHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		JSONError(c, http.StatusUnauthorized, err.Error(), err)
		return
	}

	var req model.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	if err := service.ChangePasswordService(c.Request.Context(), userID, req); err != nil {
		if errors.Is(err, service.ErrInvalidOldPassword) || errors.Is(err, service.ErrUserNotFoundOrInvalidCredentials) {
			JSONError(c, http.StatusUnauthorized, err.Error(), err)
			return
		}

		// Validation errors from utils will be returned directly (e.g., ErrorShortPassword)
		if errors.Is(err, utils.ErrorShortPassword) || errors.Is(err, utils.ErrorPasswordTooLong) || errors.Is(err, utils.ErrorInvalidCharactersInPassword) {
			JSONError(c, http.StatusBadRequest, err.Error(), err)
			return
		}

		slog.Error("Handler: ChangePassword failed unexpectedly", slog.String("userID", userID.String()), slog.Any("error", err))
		JSONError(c, http.StatusInternalServerError, "Failed to change password", err)
		return
	}

	JSONSuccess(c, http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func UpdateEmailHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		JSONError(c, http.StatusUnauthorized, err.Error(), err)
		return
	}

	var req model.UpdateEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	if err := service.UpdateEmailService(c.Request.Context(), userID, req); err != nil {
		if errors.Is(err, service.ErrInvalidPassword) || errors.Is(err, service.ErrUserNotFoundOrInvalidCredentials) {
			JSONError(c, http.StatusUnauthorized, err.Error(), err)
			return
		}

		if errors.Is(err, service.ErrEmailAlreadyInUse) || errors.Is(err, utils.ErrorInvalidEmail) {
			JSONError(c, http.StatusConflict, err.Error(), err)
			return
		}

		slog.Error("Handler: UpdateEmail failed unexpectedly", slog.String("userID", userID.String()), slog.Any("error", err))
		JSONError(c, http.StatusInternalServerError, "Failed to update email", err)
		return
	}

	JSONSuccess(c, http.StatusOK, gin.H{"message": "Email updated successfully"})
}

func DeleteAccountHandler(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		JSONError(c, http.StatusUnauthorized, err.Error(), err)
		return
	}

	var req model.DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	if err := service.DeleteAccountService(c.Request.Context(), userID, req); err != nil {
		if errors.Is(err, service.ErrInvalidPassword) || errors.Is(err, service.ErrUserNotFoundOrInvalidCredentials) {
			JSONError(c, http.StatusUnauthorized, err.Error(), err)
			return
		}

		slog.Error("Handler: DeleteAccount failed unexpectedly", slog.String("userID", userID.String()), slog.Any("error", err))
		JSONError(c, http.StatusInternalServerError, "Failed to delete account", err)
		return
	}

	JSONSuccess(c, http.StatusOK, gin.H{"message": "Account deleted successfully"})
}
