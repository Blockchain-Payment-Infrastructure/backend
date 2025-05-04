package services

import (
	"database/sql"
	"errors"
	"log"

	"github.com/Blockchain-Payment-Infrastructure/backend/auth/models"
	"github.com/Blockchain-Payment-Infrastructure/backend/db"
)

var (
	ErrUserNotFound = errors.New("User not found")
)

func UserExists(email, phone string) bool {
	var exists bool
	err := db.DB.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 OR phone = $2)", email, phone)
	if err != nil {
		log.Println("Error while checking if user exists: ", err)
	}

	return exists
}

func CreateUser(user models.User) error {
	_, err := db.DB.Exec(`INSERT INTO users (id, email, phone, hashed_password) VALUES ($1, $2, $3, $4)`, user.ID, user.Email, user.Phone, user.HashedPassword)
	if err != nil {
		log.Println("Error while creating user: ", err)
	}

	return err
}

func GetUserByEmail(email string) (models.User, error) {
	var user models.User

	err := db.DB.Get(&user, `SELECT id, email, phone, hashed_password FROM users WHERE email = $1`, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, ErrUserNotFound
		}

		return user, err
	}

	return user, nil
}
