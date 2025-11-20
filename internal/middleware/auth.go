package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"strings"
	"time"

	"backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Use the shared JwtSecretKey from the utils package so token generation and
// validation use the same secret.
var JwtSecretKey = utils.JwtSecretKey

func init() {
	if len(JwtSecretKey) == 0 {
		slog.Error("JWT_SECRET_KEY environment variable is not set!")
		JwtSecretKey = []byte("super_insecure_default_key_for_dev_only_change_this") // INSECURE DEFAULT
		slog.Warn("Using a default insecure JWT_SECRET_KEY for development. PLEASE SET JWT_SECRET_KEY environment variable.")
	}
}

// AuthMiddleware validates JWT tokens and sets the user ID in the context.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			slog.Warn("AuthMiddleware: missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check for "Bearer " prefix and extract the token
		if !strings.HasPrefix(tokenString, "Bearer ") {
			// Log a hint (but do NOT log the token itself)
			slog.Warn("AuthMiddleware: Authorization header missing 'Bearer ' prefix", slog.Int("len", len(tokenString)))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format; expected 'Authorization: Bearer <token>'"})
			c.Abort()
			return
		}
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the alg is what we expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return JwtSecretKey, nil
		})

		if err != nil {
			slog.Warn("AuthMiddleware: Token parsing failed", slog.Any("error", err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Check expiration
			if exp, ok := claims["exp"].(float64); ok {
				if int64(exp) < time.Now().Unix() {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
					c.Abort()
					return
				}
			} else {
				slog.Warn("AuthMiddleware: token missing 'exp' claim")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims (missing exp)"})
				c.Abort()
				return
			}

			// Extract user ID and set it in the context. Tokens should include user_id as a UUID string.
			if userIDStr, ok := claims["user_id"].(string); ok {
				parsedID, err := uuid.Parse(userIDStr)
				if err != nil {
					slog.Error("AuthMiddleware: invalid user_id in token claims", slog.Any("user_id", userIDStr), slog.Any("error", err))
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims (user_id malformed)"})
					c.Abort()
					return
				}
				// Store the UUID as a string in the Gin context to match handler
				// expectations (handlers read it as a string).
				c.Set(string(UserIDKey), parsedID.String())
				c.Next()
			} else {
				slog.Error("AuthMiddleware: User ID not found or invalid in token claims")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims (missing user_id)"})
				c.Abort()
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
		}
	}
}
