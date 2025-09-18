package service

import (
	"backend/internal/model"
	"backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func GetWalletAddressFromPhone(c *gin.Context, phone string) ([]model.WalletAddress, error) {
	return repository.GetWalletAddressesFromPhone(c, phone)
}
