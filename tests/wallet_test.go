package tests

import (
	"backend/internal/api/handler"
	"backend/internal/database"
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/repository"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
)

func TestWalletAPI(t *testing.T) {
	database.New(testDSN)
	database.Migrate("file://../db/migrations")

	r := gin.Default()
	r.Use(middleware.StructuredLogger())
	auth := r.Group("/auth")
	auth.POST("/signup", handler.SignUpHandler)

	wallet := r.Group("/wallet")
	wallet.Use(middleware.AuthMiddleware())
	{
		wallet.POST("/connect", handler.ConnectWalletHandler)
		wallet.GET("/addresses/:phone_number", handler.WalletAddressFromPhoneHandler)
		wallet.GET("/balance/:address", handler.GetWalletBalanceHandler)
	}

	// 1. Create a test user
	user := model.UserSignUp{
		Email:       "wallet_test@example.com",
		Username:    "walletuser",
		PhoneNumber: "1122334455",
		Password:    "TestPassword123!",
	}
	userJSON, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/auth/signup", bytes.NewBuffer(userJSON))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Failed to create user for wallet tests: %s", rr.Body.String())
	}

	// Generate a mock JWT for the user
	dbUser, err := repository.FindUserByEmail(context.Background(), user.Email)
	if err != nil {
		t.Fatalf("Could not find created user to generate JWT: %v", err)
	}
	mockToken := generateMockJWT(dbUser.ID.String())

	var connectedAddress string

	// 2. Test Wallet Connection
	t.Run("ConnectWallet", func(t *testing.T) {
		// Generate a new private key
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			t.Fatal(err)
		}
		publicKey := privateKey.PublicKey
		address := crypto.PubkeyToAddress(publicKey).Hex()

		// Sign a message
		message := "Connect wallet"
		prefixedMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
		messageHash := crypto.Keccak256Hash([]byte(prefixedMessage))
		signature, err := crypto.Sign(messageHash.Bytes(), privateKey)
		if err != nil {
			t.Fatal(err)
		}
		signature[64] += 27 // Adjust recovery ID

		connectReq := model.ConnectWalletRequest{
			Message:   message,
			Signature: hexutil.Encode(signature),
		}
		connectJSON, _ := json.Marshal(connectReq)

		req, _ := http.NewRequest("POST", "/wallet/connect", bytes.NewBuffer(connectJSON))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+mockToken)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Connect wallet failed: got %v want %v, body: %s", status, http.StatusOK, rr.Body.String())
			return
		}

		var response map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &response)
		if response["walletAddress"] != address {
			t.Errorf("Expected connected address %s, got %s", address, response["walletAddress"])
		}
		connectedAddress = address
	})

	// 3. Test Get Wallet by Phone
	t.Run("GetWalletByPhone", func(t *testing.T) {
		if connectedAddress == "" {
			t.Skip("Skipping test because wallet connection failed")
		}

		req, _ := http.NewRequest("GET", "/wallet/addresses/"+user.PhoneNumber, nil)
		req.Header.Set("Authorization", "Bearer "+mockToken)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Get wallet by phone failed: got %v want %v, body: %s", status, http.StatusOK, rr.Body.String())
			return
		}

		var addresses []model.WalletAddress
		json.Unmarshal(rr.Body.Bytes(), &addresses)
		if len(addresses) != 1 {
			t.Fatalf("Expected 1 address, got %d", len(addresses))
		}
		if addresses[0].Address != connectedAddress {
			t.Errorf("Expected address %s, got %s", connectedAddress, addresses[0].Address)
		}
	})

	// 4. Test Get Balance of Unowned Wallet
	t.Run("GetBalance_Forbidden", func(t *testing.T) {
		// Generate a random address the user does not own
		unownedAddress := "0x1234567890123456789012345678901234567890"

		req, _ := http.NewRequest("GET", "/wallet/balance/"+unownedAddress, nil)
		req.Header.Set("Authorization", "Bearer "+mockToken)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusForbidden {
			t.Errorf("Expected 403 Forbidden for unowned wallet, got %v", status)
		}
	})
}
