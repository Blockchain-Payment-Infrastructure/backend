// File: backend/internal/middleware/auth.go
package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"backend/internal/api/handler" // IMPORTANT: Changed import path to YOUR handler package

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		parts := strings.Split(tokenString, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			return handler.JwtSecretKey, nil // Use the JwtSecretKey from YOUR handler package
		})

		if err != nil || !token.Valid {
			slog.Warn("Invalid token received", slog.Any("error", err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Next()
	}
}