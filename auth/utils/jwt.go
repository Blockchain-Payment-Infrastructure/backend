package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var JWTSecret []byte

// InitialiseJWT initializes the JWT secret key.
func InitialiseJWT(key string) {
	JWTSecret = []byte(key)
}

// GenerateJWT generates a JWT token with userID and 24-hour expiration.
func GenerateJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	return token.SignedString(JWTSecret)
}
