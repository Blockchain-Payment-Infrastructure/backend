package handler

import (
	"backend/internal/ethclient"
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
)

var (
	ErrorInvalidPhoneNumber = errors.New("invalid phone number")
)

// WalletAddressFromPhoneHandler godoc
//
//	@Summary		Get Wallet Address by Phone Number
//	@Description	Retrieves wallet addresses associated with a phone number
//	@Tags			wallet
//	@Produce		json
//	@Param			phone_number	path		string					true	"Phone Number"
//	@Success		200				{array}		model.WalletAddress	"List of wallet addresses"
//	@Failure		400				{string}	string					"Invalid phone number"
//	@Router			/wallet/addresses/{phone_number} [get]
//	@Security		BearerAuth
func WalletAddressFromPhoneHandler(c *gin.Context) {
	if _, exists := c.Get(string(middleware.UserIDKey)); !exists {
		JSONError(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	phoneNumber := c.Param("phone_number")
	if len(phoneNumber) != 10 {
		JSONError(c, http.StatusBadRequest, ErrorInvalidPhoneNumber.Error(), ErrorInvalidPhoneNumber)
		return
	}

	if addresses, err := service.GetWalletAddressFromPhone(c, phoneNumber); err == nil {
		JSONSuccess(c, http.StatusOK, addresses)
	} else {
		JSONError(c, http.StatusBadRequest, err.Error(), err)
	}
}

// ConnectWalletHandler godoc
//
//	@Summary		Connect Wallet
//	@Description	Connects a user's wallet by verifying a signed message
//	@Tags			wallet
//	@Accept			json
//	@Produce		json
//	@Param			connectRequest	body		model.ConnectWalletRequest	true	"Wallet connection request"
//	@Success		200				{object}	map[string]interface{}		"Wallet connected successfully"
//	@Failure		400				{string}	string						"Invalid request payload or signature"
//	@Failure		401				{string}	string						"Unauthorized or signature verification failed"
//	@Failure		409				{string}	string						"Wallet already linked to an account"
//	@Failure		500				{string}	string						"Internal server error"
//	@Router			/wallet/connect [post]
//	@Security		BearerAuth
func ConnectWalletHandler(c *gin.Context) {
	id, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		JSONError(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userID, ok := id.(string)
	if !ok {
		JSONError(c, http.StatusInternalServerError, "Invalid user ID in context", nil)
		return
	}

	var req model.ConnectWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	prefixedMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(req.Message), req.Message)
	messageHash := crypto.Keccak256Hash([]byte(prefixedMessage))

	sig, err := hexutil.Decode(req.Signature)
	if err != nil {
		JSONError(c, http.StatusBadRequest, "Invalid signature format", err)
		return
	}

	// Adjust recovery ID
	if sig[64] == 27 || sig[64] == 28 {
		sig[64] -= 27
	}

	pubKeyBytes, err := crypto.Ecrecover(messageHash.Bytes(), sig)
	if err != nil {
		JSONError(c, http.StatusUnauthorized, "Signature verification failed", err)
		return
	}

	ecdsaPubKey, err := crypto.UnmarshalPubkey(pubKeyBytes)
	if err != nil {
		JSONError(c, http.StatusInternalServerError, "Failed to unmarshal public key", err)
		return
	}

	recoveredAddr := crypto.PubkeyToAddress(*ecdsaPubKey).Hex()

	phoneNumber, err := repository.GetPhoneNumberByUserID(c, userID)
	if err != nil {
		JSONError(c, http.StatusInternalServerError, "User's phone number not found", err)
		return
	}

	err = repository.InsertWalletAddressPhone(c, recoveredAddr, phoneNumber)
	if err != nil {
		if errors.Is(err, repository.ErrorWalletAddressAlreadyExists) {
			JSONError(c, http.StatusConflict, "This wallet is already linked to an account", err)
			return
		}
		JSONError(c, http.StatusInternalServerError, "Failed to connect wallet", err)
		return
	}

	JSONSuccess(c, http.StatusOK, gin.H{
		"success":       true,
		"walletAddress": recoveredAddr,
		"message":       "Wallet successfully connected!",
	})
}

// GetWalletBalanceHandler godoc
//
//	@Summary		Get Wallet Balance
//	@Description	Gets the ETH balance of a wallet address
//	@Tags			wallet
//	@Produce		json
//	@Param			address	path		string	true	"Wallet Address"
//	@Success		200		{object}	map[string]interface{}	"Wallet balance information"
//	@Failure		400		{string}	string	"Invalid address"
//	@Failure		401		{string}	string	"Unauthorized"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/wallet/balance/{address} [get]
//	@Security		BearerAuth
func GetWalletBalanceHandler(c *gin.Context) {
	userID, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		JSONError(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		JSONError(c, http.StatusInternalServerError, "Invalid user ID in context", nil)
		return
	}

	address := c.Param("address")
	if address == "" || len(address) != 42 {
		JSONError(c, http.StatusBadRequest, "Valid wallet address is required", nil)
		return
	}

	// Verify user owns this wallet
	userWallets, err := repository.GetUserWalletAddresses(c, userIDStr)
	if err != nil {
		slog.Error("Failed to get user wallet addresses", slog.Any("error", err))
		JSONError(c, http.StatusInternalServerError, "Failed to verify wallet ownership", err)
		return
	}

	addressOwned := false
	for _, wallet := range userWallets {
		if equalAddresses(wallet, address) {
			addressOwned = true
			break
		}
	}

	if !addressOwned {
		JSONError(c, http.StatusForbidden, "You don't own this wallet address", nil)
		return
	}

	// Get balance from blockchain
	ethClient, err := ethclient.NewClient()
	if err != nil {
		slog.Error("Failed to create Ethereum client", slog.Any("error", err))
		JSONError(c, http.StatusInternalServerError, "Failed to connect to blockchain", err)
		return
	}
	defer ethClient.Close()

	balanceWei, err := ethClient.GetETHBalance(c, address)
	if err != nil {
		slog.Error("Failed to get wallet balance", slog.Any("error", err), slog.String("address", address))
		JSONError(c, http.StatusInternalServerError, "Failed to get balance", err)
		return
	}

	balanceEther, err := ethClient.GetETHBalanceInEther(c, address)
	if err != nil {
		slog.Error("Failed to convert balance to ether", slog.Any("error", err))
		JSONError(c, http.StatusInternalServerError, "Failed to get balance", err)
		return
	}

	JSONSuccess(c, http.StatusOK, gin.H{
		"address":     address,
		"balance_wei": balanceWei.String(),
		"balance_eth": fmt.Sprintf("%.6f", balanceEther),
		"formatted":   ethclient.FormatBalanceForDisplay(balanceWei, 4),
	})
}

// GetUserWalletBalancesHandler godoc
//
//	@Summary		Get All User Wallet Balances
//	@Description	Gets ETH balances for all user's connected wallets
//	@Tags			wallet
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"All wallet balances"
//	@Failure		401	{string}	string	"Unauthorized"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/wallet/balances [get]
//	@Security		BearerAuth
func GetUserWalletBalancesHandler(c *gin.Context) {
	userID, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		JSONError(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		JSONError(c, http.StatusInternalServerError, "Invalid user ID in context", nil)
		return
	}

	// Get user's wallet addresses
	userWallets, err := repository.GetUserWalletAddresses(c, userIDStr)
	if err != nil {
		slog.Error("Failed to get user wallet addresses", slog.Any("error", err))
		JSONError(c, http.StatusInternalServerError, "Failed to get wallet addresses", err)
		return
	}

	if len(userWallets) == 0 {
		JSONSuccess(c, http.StatusOK, gin.H{
			"wallets":       []any{},
			"total_balance": "0 ETH",
		})
		return
	}

	// Get balances from blockchain
	ethClient, err := ethclient.NewClient()
	if err != nil {
		slog.Error("Failed to create Ethereum client", slog.Any("error", err))
		JSONError(c, http.StatusInternalServerError, "Failed to connect to blockchain", err)
		return
	}
	defer ethClient.Close()

	balances, err := ethClient.GetMultipleBalances(c, userWallets)
	if err != nil {
		slog.Error("Failed to get wallet balances", slog.Any("error", err))
		JSONError(c, http.StatusInternalServerError, "Failed to get balances", err)
		return
	}

	var walletBalances []map[string]any
	totalWei := balances[userWallets[0]].Uint64() * 0 // Initialize to 0

	for _, address := range userWallets {
		balance, exists := balances[address]
		if !exists {
			continue
		}

		balanceEther := ethclient.WeiToEther(balance)
		etherFloat, _ := balanceEther.Float64()

		walletBalances = append(walletBalances, map[string]any{
			"address":     address,
			"balance_wei": balance.String(),
			"balance_eth": fmt.Sprintf("%.6f", etherFloat),
			"formatted":   ethclient.FormatBalanceForDisplay(balance, 4),
		})

		// Add to total (using big.Int arithmetic would be better for precision)
		totalWei += balance.Uint64()
	}

	JSONSuccess(c, http.StatusOK, gin.H{
		"wallets":       walletBalances,
		"total_balance": fmt.Sprintf("%.6f ETH", float64(totalWei)/1e18),
		"wallet_count":  len(walletBalances),
	})
}

func equalAddresses(addr1, addr2 string) bool {
	if len(addr1) != 42 || len(addr2) != 42 {
		return false
	}
	if addr1[:2] != "0x" || addr2[:2] != "0x" {
		return false
	}
	return addr1 == addr2 // Simple comparison, could use common.HexToAddress for better normalization
}
