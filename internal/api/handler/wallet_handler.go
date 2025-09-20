package handler

import (
	"backend/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	ErrorInvalidPhoneNumber = errors.New("invalid phone number")
)

func WalletAddressFromPhoneHandler(c *gin.Context) {
	phoneNumber := c.Param("phone_number")
	if len(phoneNumber) != 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrorInvalidPhoneNumber})
	}

	if addresses, err := service.GetWalletAddressFromPhone(c, phoneNumber); err == nil {
		c.JSON(http.StatusOK, addresses)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}
