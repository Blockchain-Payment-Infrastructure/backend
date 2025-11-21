package handler

import (
	"backend/internal/config"
	"backend/internal/model"
	"backend/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SignUpHandler godoc
//
//	@Summary		User Sign Up
//	@Description	Creates a new user account after validating username, email, phoneNumber and password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			userDetails	body		model.UserSignUp	true	"User sign up details"
//	@Success		201			{string}	string				"Account created successfully"
//	@Failure		400			{string}	string				"Validation error or bad request"
//	@Failure		500			{string}	string				"Internal server error"
//	@Router			/auth/signup [post]
func SignUpHandler(c *gin.Context) {
	var userDetails model.UserSignUp
	if err := c.ShouldBindJSON(&userDetails); err != nil {
		JSONError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := service.SignUpService(c, userDetails); err != nil {
		JSONError(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	JSONSuccess(c, http.StatusCreated, gin.H{"result": "account created successfully"})
}

// LoginHandler godoc
//
//	@Summary		User Login
//	@Description	Logs the user in, returning a short-lived access token in the response and a long-lived refresh token in a secure HttpOnly cookie.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			loginDetails	body		model.UserLogin			true	"User login details"
//	@Success		200				{object}	map[string]string		"Login successful!"
//	@Failure		400				{string}	string					"Validation error"
//	@Failure		401				{string}	string					"Invalid credentials"
//	@Failure		500				{string}	string					"Internal server error"
//	@Router			/auth/login [post]
func LoginHandler(c *gin.Context) {
	var loginDetails model.UserLogin

	if err := c.ShouldBindJSON(&loginDetails); err != nil {
		JSONError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	accessToken, refreshToken, err := service.LoginService(c, loginDetails)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			JSONError(c, http.StatusUnauthorized, "Invalid credentials", err)
			return
		}
		JSONError(c, http.StatusInternalServerError, "Internal server error", err)
		return
	}

	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("refresh_token", refreshToken, 3600*24*7, "/", "", config.SecureCookie, true)

	JSONSuccess(c, http.StatusOK, gin.H{"message": "Login successful!", "access_token": accessToken})
}

// RefreshTokenHandler godoc
//
//	@Summary		Refresh Access Token
//	@Description	Generates a new access token using the refresh token sent in the HttpOnly cookie.
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	map[string]string	"New access token generated successfully"
//	@Failure		401	{string}	string				"Unauthorized or invalid refresh token"
//	@Router			/auth/refresh [post]
func RefreshTokenHandler(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		JSONError(c, http.StatusUnauthorized, "Refresh token not found", err)
		return
	}

	accessToken, err := service.RefreshTokenService(c, refreshToken)
	if err != nil {
		JSONError(c, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	JSONSuccess(c, http.StatusOK, gin.H{"access_token": accessToken})
}

// LogoutHandler godoc
//
//	@Summary		User Logout
//	@Description	Logs the user out by invalidating their refresh token and clearing the associated cookie.
//	@Tags			auth
//	@Produce		json
//	@Success		200	{string}	string	"Logout successful"
//	@Router			/auth/logout [post]
func LogoutHandler(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err == nil {
		_ = service.LogoutService(c, refreshToken)
	}

	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	JSONSuccess(c, http.StatusOK, gin.H{"message": "Logout successful"})
}
