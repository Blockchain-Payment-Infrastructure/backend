package utils

import (
	"backend/internal/config"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var JwtSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

// Predeclared errors for predictable error handling by callers.
var (
	ErrJWTSecretKeyNotConfigured = errors.New("jwt secret key is not configured")
)

func init() {
	if len(JwtSecretKey) == 0 {
		if config.AppMode == gin.DebugMode {
			JwtSecretKey = []byte("test_secret_key")
		} else {
			panic("JWT_SECRET_KEY environment variable is not set")
		}
	}
}

// GenerateAccessToken accepts a UUID for the user ID and places the UUID
// string into the token claims under `user_id`.
func GenerateAccessToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
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
