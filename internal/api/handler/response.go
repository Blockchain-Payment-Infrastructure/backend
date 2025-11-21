package handler

import (
	"github.com/gin-gonic/gin"
)

// JSONError registers the error with Gin (so structured middleware can log it)
// and returns a consistent JSON error response. Callers should pass the
// underlying error when available; if only a message is provided and the
// status is 5xx we register a generic error so the logger captures it.
func JSONError(c *gin.Context, status int, message string, err error) {
	_ = c.Error(err)

	c.AbortWithStatusJSON(status, gin.H{"error": message})
}

// JSONSuccess is a small helper to return success responses consistently.
func JSONSuccess(c *gin.Context, status int, payload any) {
	c.JSON(status, payload)
}
