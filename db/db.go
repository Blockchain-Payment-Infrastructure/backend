package db

import (
	"log"
	"net/url"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func InitDB(connectionString string) error {
	var err error
	conn, _ := url.Parse(connectionString)
	conn.RawQuery = "sslmode=verify-ca;sslrootcert=/tmp/ca.pem"
	DB, err = sqlx.Open("postgres", connectionString)
	if err != nil {
		return err
	}

	if err := DB.Ping(); err != nil {
		return err
	}

	log.Println("Database connected successfully")
	return nil
}

func CloseDB() {
	if err := DB.Close(); err != nil {
		log.Fatalf("Could not close database: %v", err)
	}
}
