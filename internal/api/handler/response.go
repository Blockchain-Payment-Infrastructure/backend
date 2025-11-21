package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// JSONError registers the error with Gin (so structured middleware can log it)
// and returns a consistent JSON error response. Callers should pass the
// underlying error when available; if only a message is provided and the
// status is 5xx we register a generic error so the logger captures it.
func JSONError(c *gin.Context, status int, message string, err error) {
	// Only register non-nil errors with Gin; c.Error(nil) panics.
	// For server errors (5xx) where callers may not provide the underlying
	// error, register a generic error so structured logging captures it.
	if err != nil {
		_ = c.Error(err)
	} else if status >= 500 {
		_ = c.Error(errors.New(message))
	}

	c.AbortWithStatusJSON(status, gin.H{"error": message})
}

// JSONSuccess is a small helper to return success responses consistently.
func JSONSuccess(c *gin.Context, status int, payload any) {
	c.JSON(status, payload)
}
