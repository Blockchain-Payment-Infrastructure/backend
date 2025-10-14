package utils

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var JwtSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

func init() {
	if len(JwtSecretKey) == 0 {
		// Check if we're in test mode
		if os.Getenv("GIN_MODE") == "test" {
			JwtSecretKey = []byte("test_secret_key")
		} else {
			slog.Error("JWT_SECRET_KEY environment variable is not set!")
			JwtSecretKey = []byte("super_insecure_default_key_for_dev_only_change_this")
			slog.Warn("Using a default insecure JWT_SECRET_KEY for development. PLEASE SET JWT_SECRET_KEY environment variable.")
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
