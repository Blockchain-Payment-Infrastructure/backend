package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/utils" // For password hashing and validation
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

<<<<<<< HEAD
// ChangePasswordService handles the business logic for changing a user's password.
func ChangePasswordService(ctx context.Context, userID uuid.UUID, req model.ChangePasswordRequest) error {
	// 1. Retrieve the user to verify the old password
	user, err := repository.FindUserByID(ctx, userID)
	if err != nil {
		slog.Error("Service: ChangePassword - User not found", slog.String("userID", userID.String()), slog.Any("error", err))
		return fmt.Errorf("user not found or invalid credentials") // Generic error for security
	}

	// 2. Verify the old password using YOUR custom utility
	match, err := utils.ComparePasswordAndHash(req.OldPassword, user.HashedPassword)
	if err != nil {
		slog.Error("Service: ChangePassword - Error comparing old password with custom util", slog.String("userID", userID.String()), slog.Any("error", err))
		return fmt.Errorf("internal server error during password verification")
	}
	if !match {
		return fmt.Errorf("invalid old password")
	}

	// 3. Validate new password using YOUR custom utility
	if valid, err := utils.ValidatePassword(req.NewPassword); !valid || err != nil {
		return err // Return specific password validation error from your utils
	}

	// 4. Hash the new password using YOUR custom utility
	newHashedPassword := utils.HashPassword(req.NewPassword)
	if newHashedPassword == "" { // Check if hashing failed (your HashPassword might return empty string on error)
		slog.Error("Service: ChangePassword - Failed to hash new password with custom util", slog.String("userID", userID.String()))
		return fmt.Errorf("failed to process new password")
	}

	// 5. Update the password in the repository
	if err := repository.UpdateUserPassword(ctx, userID, newHashedPassword); err != nil {
		slog.Error("Service: ChangePassword - Failed to update password in DB", slog.String("userID", userID.String()), slog.Any("error", err))
		return fmt.Errorf("failed to change password")
=======
var (
	ErrUserNotFoundOrInvalidCredentials = errors.New("user not found or invalid credentials")
	ErrInvalidOldPassword               = errors.New("invalid old password")
	ErrInvalidPassword                  = errors.New("invalid password")
	ErrInternalPasswordVerification     = errors.New("internal server error during password verification")
	ErrEmailAlreadyInUse                = errors.New("email already in use by another account")
	ErrFailedToUpdateEmail              = errors.New("failed to update email")
	ErrFailedToChangePassword           = errors.New("failed to change password")
	ErrFailedToDeleteAccount            = errors.New("failed to delete account")
)

func ChangePasswordService(ctx context.Context, userID uuid.UUID, req model.UpdatePasswordRequest) error {
	user, err := repository.FindUserByID(ctx, userID)
	if err != nil {
		slog.Error("Service: ChangePassword - User not found", slog.String("userID", userID.String()), slog.Any("error", err))
		return ErrUserNotFoundOrInvalidCredentials
	}

	match, err := utils.ComparePasswordAndHash(req.OldPassword, user.HashedPassword)
	if err != nil {
		slog.Error("Service: ChangePassword - Error comparing old password with custom util", slog.String("userID", userID.String()), slog.Any("error", err))
		return ErrInternalPasswordVerification
	}
	if !match {
		return ErrInvalidOldPassword
	}

	if valid, err := utils.ValidatePassword(req.NewPassword); !valid || err != nil {
		return err
	}

	if err := repository.UpdateUserPassword(ctx, userID, utils.HashPassword(req.NewPassword)); err != nil {
		slog.Error("Service: ChangePassword - Failed to update password in DB", slog.String("userID", userID.String()), slog.Any("error", err))
		return ErrFailedToChangePassword
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
	}

	slog.Info("Service: Password changed successfully", slog.String("userID", userID.String()))
	return nil
}

// UpdateEmailService handles the business logic for updating a user's email.
func UpdateEmailService(ctx context.Context, userID uuid.UUID, req model.UpdateEmailRequest) error {
	// 1. Retrieve the user to verify the password
	user, err := repository.FindUserByID(ctx, userID)
	if err != nil {
		slog.Error("Service: UpdateEmail - User not found", slog.String("userID", userID.String()), slog.Any("error", err))
<<<<<<< HEAD
		return fmt.Errorf("user not found or invalid credentials") // Generic error for security
=======
		return ErrUserNotFoundOrInvalidCredentials
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
	}

	// 2. Verify the current password using YOUR custom utility
	match, err := utils.ComparePasswordAndHash(req.Password, user.HashedPassword)
	if err != nil {
		slog.Error("Service: UpdateEmail - Error comparing password with custom util", slog.String("userID", userID.String()), slog.Any("error", err))
<<<<<<< HEAD
		return fmt.Errorf("internal server error during password verification")
	}
	if !match {
		return fmt.Errorf("invalid password")
=======
		return ErrInternalPasswordVerification
	}
	if !match {
		return ErrInvalidPassword
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
	}

	// 3. Validate new email (format handled by Gin binding in handler, but can add more here if needed)
	if err := utils.ValidateEmail(req.NewEmail); err != nil {
		return err // Return specific email validation error
	}

	// Optional: Check if the new email is already in use by another user
	existingUser, err := repository.FindUserByEmail(ctx, req.NewEmail)
	if err == nil && existingUser != nil && existingUser.ID != userID {
<<<<<<< HEAD
		return fmt.Errorf("email already in use by another account")
	}
	if err != nil {
		// Accept either the exact sentinel error or an equivalent error string
		if errors.Is(err, repository.ErrorUserNotFound) || err.Error() == repository.ErrorUserNotFound.Error() {
			// email not found — OK to proceed
		} else {
			// Some DB error occurred
=======
		return ErrEmailAlreadyInUse
	}
	if err != nil {
		// If error is anything other than user-not-found, treat as DB error
		if !errors.Is(err, repository.ErrorUserNotFound) {
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
			slog.Error("Service: UpdateEmail - Error checking for existing email", slog.Any("error", err))
			return fmt.Errorf("failed to check email uniqueness")
		}
	}

	// 4. Update the email in the repository
	if err := repository.UpdateUserEmail(ctx, userID, req.NewEmail); err != nil {
		slog.Error("Service: UpdateEmail - Failed to update email in DB", slog.String("userID", userID.String()), slog.Any("error", err))
<<<<<<< HEAD
		return fmt.Errorf("failed to update email")
=======
		return ErrFailedToUpdateEmail
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
	}

	slog.Info("Service: Email updated successfully", slog.String("userID", userID.String()), slog.String("newEmail", req.NewEmail))
	return nil
}

// DeleteAccountService handles the business logic for deleting a user's account.
func DeleteAccountService(ctx context.Context, userID uuid.UUID, req model.DeleteAccountRequest) error {
	// 1. Retrieve the user to verify the password
	user, err := repository.FindUserByID(ctx, userID)
	if err != nil {
		slog.Error("Service: DeleteAccount - User not found", slog.String("userID", userID.String()), slog.Any("error", err))
<<<<<<< HEAD
		return fmt.Errorf("user not found or invalid credentials")
=======
		return ErrUserNotFoundOrInvalidCredentials
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
	}

	// 2. Verify the current password using YOUR custom utility
	match, err := utils.ComparePasswordAndHash(req.Password, user.HashedPassword)
	if err != nil {
		slog.Error("Service: DeleteAccount - Error comparing password with custom util", slog.String("userID", userID.String()), slog.Any("error", err))
<<<<<<< HEAD
		return fmt.Errorf("internal server error during password verification")
	}
	if !match {
		return fmt.Errorf("invalid password")
=======
		return ErrInternalPasswordVerification
	}
	if !match {
		return ErrInvalidPassword
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
	}

	// 3. Delete the user and associated data in the repository (transactional)
	if err := repository.DeleteUser(ctx, userID); err != nil {
		slog.Error("Service: DeleteAccount - Failed to delete user in DB", slog.String("userID", userID.String()), slog.Any("error", err))
<<<<<<< HEAD
		return fmt.Errorf("failed to delete account")
=======
		return ErrFailedToDeleteAccount
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
	}

	slog.Info("Service: Account deleted successfully", slog.String("userID", userID.String()))
	return nil
}
