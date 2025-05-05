package config

import (
	"log"
	"os"
	"path/filepath"
)

var (
	DatabaseURL string
	JWTSecret   string
	AivenPGCert string
)

func LoadEnv() error {
	DatabaseURL = os.Getenv("DATABASE_URL")
	if DatabaseURL == "" {
		log.Fatal("DATABASE_URL is not set in environment")
	}

	JWTSecret = os.Getenv("JWT_ACCESS_SECRET")
	if JWTSecret == "" {
		log.Fatal("JWT_ACCESS_SECRET is not set in environment")
	}

	AivenPGCert = os.Getenv("AIVEN_PG_CERT")
	if AivenPGCert == "" {
		log.Fatal("AIVEN_PG_CERT is not set in environment")
	}
	if err := writeCertToFile(); err != nil {
		log.Fatalf("Unable to write AIVEN_PG_CERT to disk: %v", err)
	}

	return nil
}

func writeCertToFile() error {
	tmpDir := os.TempDir()
	certPath := filepath.Join(tmpDir, "ca.pem")
	err := os.WriteFile(certPath, []byte(AivenPGCert), 0600)

	return err
}
