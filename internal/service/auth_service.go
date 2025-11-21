package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/utils"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
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

func LoginService(c *gin.Context, loginDetails model.UserLogin) (string, string, error) {
	user, err := repository.FindUserByEmail(c.Request.Context(), loginDetails.Email)
	if err != nil {
		if errors.Is(err, repository.ErrorUserNotFound) {
			slog.Warn("Login failed for email (user not found)", slog.String("email", loginDetails.Email))
			return "", "", ErrInvalidCredentials
		}

		slog.Error("Login failed due to database error", slog.String("email", loginDetails.Email), slog.Any("error", err))
		return "", "", err
	}

	match, err := utils.ComparePasswordAndHash(loginDetails.Password, user.HashedPassword)
	if err != nil {
		slog.Error("Error comparing password and hash", slog.Any("error", err))
		return "", "", err
	}

	if !match {
		slog.Warn("Login failed for email (password mismatch)", slog.String("email", loginDetails.Email))
		return "", "", ErrInvalidCredentials
	}

	accessToken, err := utils.GenerateAccessToken(user.ID)
	if err != nil {
		slog.Error("Error generating access token", slog.Any("error", err))
		return "", "", err
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		slog.Error("Error generating refresh token", slog.Any("error", err))
		return "", "", err
	}

	refreshTokenHash := repository.HashRefreshToken(refreshToken)
	expiresAt := time.Now().Add(time.Hour * 24 * 7) // 7-day expiry

	if err := repository.StoreRefreshToken(c.Request.Context(), user.ID, refreshTokenHash, expiresAt); err != nil {
		slog.Error("Failed to store refresh token", slog.Any("error", err))
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func RefreshTokenService(c *gin.Context, refreshToken string) (string, error) {
	refreshTokenHash := repository.HashRefreshToken(refreshToken)

	userID, err := repository.GetUserByRefreshToken(c.Request.Context(), refreshTokenHash)
	if err != nil {
		return "", err
	}

	accessToken, err := utils.GenerateAccessToken(userID)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func LogoutService(c *gin.Context, refreshToken string) error {
	return repository.DeleteRefreshToken(c.Request.Context(), repository.HashRefreshToken(refreshToken))
}
