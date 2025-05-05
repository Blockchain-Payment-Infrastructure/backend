package handlers

import (
	"errors"
	"log"
	"net/http"
	"net/mail"

	"github.com/Blockchain-Payment-Infrastructure/backend/auth/models"
	"github.com/Blockchain-Payment-Infrastructure/backend/auth/services"
	"github.com/Blockchain-Payment-Infrastructure/backend/auth/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	ErrInvalidEmailAddress = errors.New("Invalid email address")
	ErrUserExists          = errors.New("User with the email or phone exists")
	ErrHashFailed          = errors.New("Failed to hash password")
	ErrCreateUserFailed    = errors.New("User creation failed")
)

func RegisterHandler(c *gin.Context) {
	var req models.RegisterRequest
	// parse request into the expected data model
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := mail.ParseAddress(req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidEmailAddress.Error()})
	}

	if services.UserExists(req.Username, req.Email, req.Phone) {
		c.JSON(http.StatusConflict, gin.H{"error": ErrUserExists.Error()})
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrHashFailed.Error()})
		return
	}

	user := models.User{
		ID:             uuid.New().String(),
		Email:          req.Email,
		Phone:          req.Phone,
		HashedPassword: hashedPassword,
	}

	err = services.CreateUser(user)
	if err != nil {
		log.Printf("User creation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrCreateUserFailed.Error()})
		return
	}
}
