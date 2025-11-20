package utils

import (
	"backend/internal/config"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var JwtSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

func init() {
	if len(JwtSecretKey) == 0 {
		// Check if we're in test mode
		if config.AppMode == gin.DebugMode {
			JwtSecretKey = []byte("test_secret_key")
		} else {
			slog.Error("JWT_SECRET_KEY environment variable is not set!")
			panic("No JWT_SECRET_KEY set")
		}
	}
}

func GenerateAccessToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Minute * 15).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(JwtSecretKey)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
