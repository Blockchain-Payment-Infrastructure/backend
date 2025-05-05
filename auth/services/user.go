package services

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/Blockchain-Payment-Infrastructure/backend/auth/models"
	"github.com/Blockchain-Payment-Infrastructure/backend/db"
)

var (
	ErrUserNotFound = errors.New("User not found")
)

func UserExists(username, email, phone string) bool {
	var exists bool
	err := db.DB.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 OR phone = $2 OR username = $3)", email, phone, username)
	if err != nil {
		log.Println("Error while checking if user exists: ", err)
	}

	return exists
}

func CreateUser(user models.User) error {
	_, err := db.DB.Exec(`INSERT INTO users (id, username, email, phone, hashed_password, created_at) VALUES ($1, $2, $3, $4, $5, $6)`, user.ID, user.Username, user.Email, user.Phone, user.HashedPassword, time.Now())
	if err != nil {
		log.Println("Error while creating user: ", err)
	}

	return err
}

func GetUserByEmail(email string) (models.User, error) {
	var user models.User

	err := db.DB.Get(&user, `SELECT id, username, email, phone, hashed_password FROM users WHERE email = $1`, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, ErrUserNotFound
		}

		return user, err
	}

	return user, nil
}

func GetUserByID(id string) (models.User, error) {
	var user models.User

	err := db.DB.Get(&user, `SELECT id, username, email, phone, hashed_password FROM users WHERE id = $1`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, ErrUserNotFound
		}

		return user, err
	}

	return user, nil
}
