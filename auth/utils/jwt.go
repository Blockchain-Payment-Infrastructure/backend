package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var JWTSecret []byte

func InitialiseJWT(access string) {
	JWTSecret = []byte(access)
}

func GenerateAccessToken(userID string) (string, error) {
	return generateJWT(userID, JWTSecret, 15*time.Minute)
}

func GenerateRefreshToken(userID string) (string, error) {
	return generateJWT(userID, JWTSecret, 15*time.Minute)
}

func generateJWT(userID string, secret []byte, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"aud": "user",
		"exp": time.Now().Add(duration).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	return token.SignedString(secret)
}
