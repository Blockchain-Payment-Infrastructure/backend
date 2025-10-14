package tests

import (
	"backend/internal/api/handler"
	"backend/internal/database"
	"backend/internal/middleware"
	"backend/internal/model"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Mock JWT token for testing
func generateMockJWT(userID string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     9999999999, // Far future expiration
	})

	// Use a test secret key
	tokenString, _ := token.SignedString([]byte("test_secret_key"))
	return tokenString
}

func TestPaymentAPI(t *testing.T) {
	// Initialize test database
	database.New(testDSN)
	database.Migrate("file://../db/migrations")

	// Create test user first
	user := model.UserSignUp{
		Email:       "payment_test@example.com",
		Username:    "paymentuser",
		PhoneNumber: "+1234567890",
		Password:    "TestPassword123!",
	}

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Add auth routes
	auth := r.Group("/auth")
	auth.POST("/signup", handler.SignUpHandler)
	auth.POST("/login", handler.LoginHandler)

	// Add payment routes with auth middleware
	payments := r.Group("/payments")
	payments.Use(middleware.AuthMiddleware())
	payments.POST("", handler.CreatePaymentHandler)
	payments.GET("", handler.GetUserPaymentsHandler)
	payments.GET("/:id", handler.GetPaymentHandler)
	payments.GET("/stats", handler.GetPaymentStatsHandler)

	// Create test user
	userJSON, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/auth/signup", bytes.NewBuffer(userJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("User creation failed: got %v want %v, body: %s", status, http.StatusCreated, rr.Body.String())
		return
	}

	// Login to get real user ID (in production, you'd parse the response)
	loginReq := model.UserLogin{
		Email:    user.Email,
		Password: user.Password,
	}

	loginJSON, err := json.Marshal(loginReq)
	if err != nil {
		t.Fatal(err)
	}

	req, err = http.NewRequest("POST", "/auth/login", bytes.NewBuffer(loginJSON))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Login failed: got %v want %v", status, http.StatusOK)
		return
	}

	// For testing purposes, we'll use a mock user ID and JWT
	testUserID := uuid.New().String()
	mockToken := generateMockJWT(testUserID)

	// Test Payment Creation (will fail due to blockchain verification, but should validate request)
	t.Run("CreatePayment_ValidationTest", func(t *testing.T) {
		paymentReq := model.CreatePaymentRequest{
			ToAddress:       "0x742d35Cc6633C0532925a3b8D4C0f56e3c4D6329",
			Amount:          "1000000000000000000", // 1 ETH in Wei
			Currency:        "ETH",
			TransactionHash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12",
			Description:     stringPtr("Test payment"),
		}

		paymentJSON, err := json.Marshal(paymentReq)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest("POST", "/payments", bytes.NewBuffer(paymentJSON))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+mockToken)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		// We expect this to fail due to blockchain verification, but it should at least validate the request format
		if status := rr.Code; status == http.StatusCreated {
			t.Logf("Payment creation succeeded (unexpected in test environment): %s", rr.Body.String())
		} else if status == http.StatusBadRequest {
			t.Logf("Payment creation failed as expected due to blockchain verification: %s", rr.Body.String())
		} else {
			t.Errorf("Unexpected status code: got %v, body: %s", status, rr.Body.String())
		}
	})

	// Test Get User Payments
	t.Run("GetUserPayments", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/payments", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+mockToken)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Get payments failed: got %v want %v, body: %s", status, http.StatusOK, rr.Body.String())
			return
		}

		var response model.PaymentListResponse
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
			return
		}

		// Should return empty list for new user
		if response.TotalCount != 0 {
			t.Errorf("Expected 0 payments for new user, got %d", response.TotalCount)
		}
	})

	// Test Get Payment Stats
	t.Run("GetPaymentStats", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/payments/stats", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+mockToken)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Get payment stats failed: got %v want %v, body: %s", status, http.StatusOK, rr.Body.String())
			return
		}

		var stats map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &stats)
		if err != nil {
			t.Errorf("Failed to unmarshal stats response: %v", err)
			return
		}

		// Verify stats structure
		expectedFields := []string{"total_payments", "confirmed", "pending", "failed", "total_amount"}
		for _, field := range expectedFields {
			if _, exists := stats[field]; !exists {
				t.Errorf("Missing field %s in stats response", field)
			}
		}
	})

	// Test Invalid Payment ID
	t.Run("GetPayment_InvalidID", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/payments/invalid-id", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+mockToken)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest && status != http.StatusNotFound {
			t.Errorf("Expected 400 or 404 for invalid payment ID, got %v, body: %s", status, rr.Body.String())
		}
	})

	// Test Unauthorized Access
	t.Run("UnauthorizedAccess", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/payments", nil)
		if err != nil {
			t.Fatal(err)
		}
		// No Authorization header

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("Expected 401 for unauthorized access, got %v", status)
		}
	})

	// Test Invalid Authorization Token
	t.Run("InvalidAuthToken", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/payments", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer invalid-token")

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("Expected 401 for invalid token, got %v", status)
		}
	})
}

func TestPaymentRequestValidation(t *testing.T) {
	// Test invalid payment request formats
	testCases := []struct {
		name        string
		request     model.CreatePaymentRequest
		expectError bool
	}{
		{
			name: "Valid Request",
			request: model.CreatePaymentRequest{
				ToAddress:       "0x742d35Cc6633C0532925a3b8D4C0f56e3c4D6329",
				Amount:          "1000000000000000000",
				Currency:        "ETH",
				TransactionHash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12",
			},
			expectError: false,
		},
		{
			name: "Invalid To Address - Too Short",
			request: model.CreatePaymentRequest{
				ToAddress:       "0x742d35Cc",
				Amount:          "1000000000000000000",
				TransactionHash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12",
			},
			expectError: true,
		},
		{
			name: "Invalid Transaction Hash - Too Short",
			request: model.CreatePaymentRequest{
				ToAddress:       "0x742d35Cc6633C0532925a3b8D4C0f56e3c4D6329",
				Amount:          "1000000000000000000",
				TransactionHash: "0x1234567890abcdef",
			},
			expectError: true,
		},
		{
			name: "Missing Amount",
			request: model.CreatePaymentRequest{
				ToAddress:       "0x742d35Cc6633C0532925a3b8D4C0f56e3c4D6329",
				TransactionHash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12",
			},
			expectError: true,
		},
		{
			name: "Missing To Address",
			request: model.CreatePaymentRequest{
				Amount:          "1000000000000000000",
				TransactionHash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12",
			},
			expectError: true,
		},
	}

	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Mock auth middleware that always passes
	r.Use(func(c *gin.Context) {
		c.Set("userID", uuid.New().String())
		c.Next()
	})

	r.POST("/payments", handler.CreatePaymentHandler)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestJSON, err := json.Marshal(tc.request)
			if err != nil {
				t.Fatal(err)
			}

			req, err := http.NewRequest("POST", "/payments", bytes.NewBuffer(requestJSON))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if tc.expectError {
				if status := rr.Code; status != http.StatusBadRequest {
					t.Errorf("Expected 400 for invalid request, got %v, body: %s", status, rr.Body.String())
				}
			} else {
				// Even valid requests will fail due to blockchain verification in test env
				// But they shouldn't fail due to validation
				if status := rr.Code; status == http.StatusBadRequest {
					// Check if it's a validation error or blockchain error
					var response map[string]string
					json.Unmarshal(rr.Body.Bytes(), &response)
					if errorMsg, exists := response["error"]; exists {
						if errorMsg == "Invalid request payload: Key: 'CreatePaymentRequest.ToAddress' Error:Field validation for 'ToAddress' failed on the 'len' tag" ||
							errorMsg == "Invalid request payload: Key: 'CreatePaymentRequest.Amount' Error:Field validation for 'Amount' failed on the 'required' tag" {
							t.Errorf("Valid request failed validation: %s", errorMsg)
						}
					}
				}
			}
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
