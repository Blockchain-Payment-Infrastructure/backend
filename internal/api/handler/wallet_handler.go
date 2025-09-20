package handler

import (
	"backend/internal/repository"
	"backend/internal/service"
	"context"
	"encoding/json"
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

type ConnectWalletRequest struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

func getUserIdFromContext(ctx context.Context) int {
	if id, ok := ctx.Value("userID").(int); ok {
		return id
	}
	return 0
}
func ConnectWalletHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIdFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ConnectWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	prefixedMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(req.Message), req.Message)
	messageHash := crypto.Keccak256Hash([]byte(prefixedMessage))
	sig, err := hexutil.Decode(req.Signature)
	if err != nil {
		http.Error(w, "Invalid signature format", http.StatusBadRequest)
		return
	}
	if sig[64] == 27 || sig[64] == 28 {
		sig[64] -= 27
	}

	pubKey, err := crypto.Ecrecover(messageHash.Bytes(), sig)
	if err != nil {
		http.Error(w, "Signature verification failed", http.StatusUnauthorized)
		return
	}

	ecdsaPubKey, err := crypto.UnmarshalPubkey(pubKey)
	if err != nil {
		http.Error(w, "Failed to unmarshal public key", http.StatusInternalServerError)
		return
	}
	recoveredAddr := crypto.PubkeyToAddress(*ecdsaPubKey).Hex()
	phoneNumber, err := repository.GetPhoneNumberByUserID(userID)
	if err != nil {
		// Handle case where phone number can't be found
		http.Error(w, "User's phone number not found", http.StatusInternalServerError)
		return
	}
	err = repository.InsertWalletAddressPhone(recoveredAddr, phoneNumber)
	if err != nil {
		if err.Error() == "wallet address already exists" {
			http.Error(w, "This wallet is already linked to an account", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to connect wallet", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"walletAddress": recoveredAddr,
		"message":       "Wallet successfully connected!",
	})
}
