package handler

import (
	"backend/internal/model"
	"backend/internal/service"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func WalletAddressFromPhoneHandler(c *gin.Context) {
	var phoneNumber model.PhoneNumber

	if err := c.BindJSON(&phoneNumber); err != nil {
		slog.Error("Binding error", slog.Any("error", err.Error()))
		return
	}

	if addresses, err := service.GetWalletAddressFromPhone(c, phoneNumber.Phone); err == nil {
		c.JSON(http.StatusOK, addresses)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}
