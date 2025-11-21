package middleware

import (
	"errors"
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

var (
	JwtSecretKey                 = utils.JwtSecretKey
	ErrorUnexpectedSigningMethod = errors.New("unexpected signing method")
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			slog.Warn("AuthMiddleware: missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		if !strings.HasPrefix(tokenString, "Bearer ") {
			slog.Warn("AuthMiddleware: Authorization header missing 'Bearer ' prefix", slog.Int("len", len(tokenString)))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format; expected 'Authorization: Bearer <token>'"})
			c.Abort()
			return
		}

		tokenString = tokenString[7:]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("%w: %v", ErrorUnexpectedSigningMethod, token.Header["alg"])
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

			if userIDStr, ok := claims["user_id"].(string); ok {
				parsedID, err := uuid.Parse(userIDStr)
				if err != nil {
					slog.Error("AuthMiddleware: invalid user_id in token claims", slog.Any("user_id", userIDStr), slog.Any("error", err))
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims (user_id malformed)"})
					c.Abort()
					return
				}

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
