package utils

import (
	"errors"
	"regexp"
)

var (
	ErrorInvalidEmail    = errors.New("invalid email format")
	ErrorInvalidUsername = errors.New("username can only contain letters, numbers, and underscores")
)

// ValidateEmail checks if the provided string is a valid email format.
func ValidateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return ErrorInvalidEmail
	}
	return nil
}

func ValidateUsername(username string) error {
	// Username must be between 3 and 20 characters
	if len(username) < 3 || len(username) > 20 {
		return errors.New("username must be between 3 and 20 characters")
	}

	// Username can only contain alphanumeric characters and underscores
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(username) {
		return ErrorInvalidUsername
	}
	return nil
}
