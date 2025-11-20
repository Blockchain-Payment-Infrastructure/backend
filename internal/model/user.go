// File: backend/internal/model/user.go (MODIFIED: Added request structs for account settings)
package model

import "github.com/google/uuid"

// UserSignUp represents the input structure for a new user registration.
// It directly maps to the JSON request body.
type UserSignUp struct {
	Email       string `json:"email" binding:"required,email"`
	Username    string `json:"username" binding:"required,min=3"`
	Password    string `json:"password" binding:"required,min=8"`
	PhoneNumber string `json:"phone_number" binding:"required,e164"` // Required, E.164 format validation
}

// UserLogin represents the input structure for user login.
type UserLogin struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// User represents the structure of a user as stored in the database.
// ID is a UUID to be consistent across the application.
type User struct {
	ID             uuid.UUID `json:"id"`
	Email          string    `json:"email"`
	Username       string    `json:"username"`
	PhoneNumber    string    `json:"phone_number"`
	HashedPassword string    `json:"-"` // Store the hashed password, omit from JSON output
}

// ChangePasswordRequest represents the input for changing a user's password.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// UpdateEmailRequest represents the input for updating a user's email.
type UpdateEmailRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
	Password string `json:"password" binding:"required"` // Current password for verification
}

// DeleteAccountRequest represents the input for deleting a user's account.
type DeleteAccountRequest struct {
	Password string `json:"password" binding:"required"` // Current password for verification
}
