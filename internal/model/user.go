// File: backend/internal/model/user.go (MODIFIED: Added 'User' struct)
package model

// UserSignUp represents the input structure for a new user registration.
// It directly maps to the JSON request body.
type UserSignUp struct {
	Email    string `json:"email" binding:"required,email"` // Added email validation
	Username string `json:"username" binding:"required,min=3"` // Added min length
	Password string `json:"password" binding:"required,min=8"` // Added min length for password
}

// User represents the structure of a user as stored in the database.
// This is separate from UserSignUp to handle the hashed password and ID.
type User struct {
	ID             uint   `json:"id"`
	Email          string `json:"email"`
	Username       string `json:"username"`
	HashedPassword string `json:"-"` // Store the hashed password, omit from JSON output
}