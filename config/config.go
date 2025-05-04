package config

import (
	"errors"
	"log"
	"os"
)

var (
	DatabaseURL string
	JWTSecret   string
)

func LoadEnv() error {
	DatabaseURL = os.Getenv("DATABASE_URL")
	if DatabaseURL == "" {
		log.Fatal("DATABASE_URL is not set in environment")
		return errors.New("DATABASE_URL is not set in environment")
	}

	JWTSecret = os.Getenv("JWT_SECRET")
	if JWTSecret == "" {
		log.Fatal("JWT_SECRET is not set in environment")
		return errors.New("JWT_SECRET is not set in environment")
	}

	return nil
}
