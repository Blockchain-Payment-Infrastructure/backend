package handler

import (
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"
	"errors"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
)

var (
	ErrorInvalidPhoneNumber = errors.New("invalid phone number")
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

func ConnectWalletHandler(c *gin.Context) {
	id, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := id.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	var req model.ConnectWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	prefixedMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(req.Message), req.Message)
	messageHash := crypto.Keccak256Hash([]byte(prefixedMessage))

	sig, err := hexutil.Decode(req.Signature)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature format"})
		return
	}

	// Adjust recovery ID
	if sig[64] == 27 || sig[64] == 28 {
		sig[64] -= 27
	}

	pubKeyBytes, err := crypto.Ecrecover(messageHash.Bytes(), sig)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Signature verification failed"})
		return
	}

	ecdsaPubKey, err := crypto.UnmarshalPubkey(pubKeyBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal public key"})
		return
	}

	recoveredAddr := crypto.PubkeyToAddress(*ecdsaPubKey).Hex()

	phoneNumber, err := repository.GetPhoneNumberByUserID(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User's phone number not found"})
		return
	}

	err = repository.InsertWalletAddressPhone(c, recoveredAddr, phoneNumber)
	if err != nil {
		if err.Error() == "wallet address already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "This wallet is already linked to an account"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect wallet"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"walletAddress": recoveredAddr,
		"message":       "Wallet successfully connected!",
	})
}
