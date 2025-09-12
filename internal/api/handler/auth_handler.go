package handler

import (
	"backend/internal/model"
	"backend/internal/service"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SignUpHandler godoc
//
//	@Summary		User Sign Up
//	@Description	Creates a new user account after validating username, email, and password
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.SignUpService(c, userDetails); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"result": "account created successfully"})
}

// LoginHandler godoc
//
//	@Summary		User Login
//	@Description	Logs the user in and returns a jwt token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			loginDetails	body		model.UserLogin	true	"User login details"
//	@Success		200				{string}	string			"Login successful!"
//	@Failure		400				{string}	string			"Validation error"
//	@Failure		500				{string}	string			"Internal server error"
//	@Router			/auth/login [post]
func LoginHandler(c *gin.Context) {
	var loginDetails model.UserLogin

	if err := c.ShouldBindJSON(&loginDetails); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	token, err := service.LoginService(c, loginDetails)
	if err != nil {
		slog.Error(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login successful!", "token": token})
}
