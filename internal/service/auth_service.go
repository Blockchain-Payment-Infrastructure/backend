package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/utils"
	"log/slog"
	"net/http"
	"strings"

	"github.com/alexedwards/argon2id"
	"github.com/gin-gonic/gin"
)

func SignUpService(c *gin.Context, userDetails model.UserSignUp) error {
	// Trim unneccessary spaces at either ends of the string
	userDetails.Password = strings.Trim(userDetails.Password, " ")
	if valid, err := utils.ValidatePassword(userDetails.Password); !valid && err != nil {
		return err
	}

	if err := utils.ValidateEmail(userDetails.Email); err != nil {
		return err
	}

	if err := utils.ValidateUsername(userDetails.Username); err != nil {
		return err
	}

	// Hash the password and insert the user
	userDetails.Password = utils.HashPassword(userDetails.Password)
	if err := repository.CreateUser(c.Request.Context(), userDetails); err != nil {
		return err
	}

	return nil
}

func LoginService(c *gin.Context, loginDetails model.UserLogin) (string, error) {
	user, err := repository.FindUserByEmail(loginDetails.Email)
	if err != nil {
		slog.Warn("Login failed for email (user not found/db error)", slog.String("email", loginDetails.Email), slog.Any("error", err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return "", err
	}

	match, err := argon2id.ComparePasswordAndHash(loginDetails.Password, user.HashedPassword)
	if err != nil {
		// Handle error (e.g., hash format is invalid, unexpected issue)
		slog.Error("Error comparing password and hash", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error during password verification"})
		return "", err
	}

	if !match {
		slog.Warn("Login failed for email (password mismatch)", slog.String("email", loginDetails.Email))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return "", err
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		slog.Error("Error generating token", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return "", err
	}

	return token, nil
}
