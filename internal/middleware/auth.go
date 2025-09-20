package middleware

import (
	"backend/internal/utils"
	"log/slog"
	"net/http"
	"strings"

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

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (any, error) {
			return utils.JwtSecretKey, nil
		})

		if err != nil || !token.Valid {
			slog.Warn("Invalid token received", slog.Any("error", err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Example: get user_id
		if userID, exists := claims["user_id"].(string); exists {
			c.Set("userID", userID)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id missing in token"})
			c.Abort()
			return
		}
		c.Next()
	}
}
