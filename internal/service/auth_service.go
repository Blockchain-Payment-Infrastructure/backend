package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/utils"
	"strings"

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
