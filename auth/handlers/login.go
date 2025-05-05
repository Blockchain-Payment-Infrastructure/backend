package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/Blockchain-Payment-Infrastructure/backend/auth/models"
	"github.com/Blockchain-Payment-Infrastructure/backend/auth/services"
	"github.com/Blockchain-Payment-Infrastructure/backend/auth/utils"
	"github.com/gin-gonic/gin"
)

var (
	ErrUserNotFound               = errors.New("User not found")
	ErrPasswordVerificationFailed = errors.New("Password verification failed")
	ErrInvalidPassword            = errors.New("Invalid password")
	ErrJWTTokenGenerationFailed   = errors.New("Failed to generate JWT token")
)

func LoginHandler(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := services.GetUserByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": ErrUserNotFound.Error()})
		return
	}

	isValid, err := utils.VerifyPassword(user.HashedPassword, req.Password)
	if err != nil {
		log.Printf("Password verification failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrPasswordVerificationFailed.Error()})
		return
	}

	if !isValid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidPassword.Error()})
		return
	}

	accessToken, err := utils.GenerateAccessToken(user.ID)
	if err != nil {
		log.Printf("Failed to generate JWT Access token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrJWTTokenGenerationFailed.Error()})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		log.Printf("Failed to generate JWT Refresh token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrJWTTokenGenerationFailed.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Login Successful",
		"refresh_token": refreshToken,
		"access_token":  accessToken,
	})
}
