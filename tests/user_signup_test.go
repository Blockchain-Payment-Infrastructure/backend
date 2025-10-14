package tests

import (
	"backend/internal/api/handler"
	"backend/internal/database"
	"backend/internal/model"
	"backend/internal/repository"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUserSignup(t *testing.T) {
	database.New(testDSN)
	database.Migrate("file://../db/migrations")

	r := gin.Default()
	auth := r.Group("/auth")
	auth.POST("/signup", handler.SignUpHandler)

	user := model.UserSignUp{
		Email:       "testing@abcd.com",
		Username:    "testing",
		PhoneNumber: "+9876543210",
		Password:    "H*mUhZ655mJo$$@K",
	}

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
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	expected := `{"result":"account created successfully"}`
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	ctx := context.Background()
	exists, err := repository.UserExists(ctx, user.Email)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("User not found in database")
	}
}

func TestUserSignup_Duplicate(t *testing.T) {
	database.New(testDSN)
	database.Migrate("file://../db/migrations")

	r := gin.Default()
	auth := r.Group("/auth")
	auth.POST("/signup", handler.SignUpHandler)

	baseUser := model.UserSignUp{
		Email:       "base@example.com",
		Username:    "baseuser",
		PhoneNumber: "1234567890",
		Password:    "TestPassword123!",
	}

	// Create the initial user that we will try to duplicate
	userJSON, _ := json.Marshal(baseUser)
	req, _ := http.NewRequest("POST", "/auth/signup", bytes.NewBuffer(userJSON))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Initial user creation failed, cannot proceed with duplicate tests: %s", rr.Body.String())
	}

	t.Run("Duplicate Email", func(t *testing.T) {
		duplicateUser := model.UserSignUp{
			Email:       baseUser.Email, // Duplicate email
			Username:    "new_username_1",
			PhoneNumber: "1000000001",
			Password:    "TestPassword123!",
		}
		userJSON, _ := json.Marshal(duplicateUser)
		req, _ := http.NewRequest("POST", "/auth/signup", bytes.NewBuffer(userJSON))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status 400 for duplicate email, got %v", status)
		}
		var response map[string]string
		json.Unmarshal(rr.Body.Bytes(), &response)
		if response["error"] != repository.ErrorEmailExists.Error() {
			t.Errorf("Expected error '%s', got '%s'", repository.ErrorEmailExists.Error(), response["error"])
		}
	})

	t.Run("Duplicate Username", func(t *testing.T) {
		duplicateUser := model.UserSignUp{
			Email:       "new_email_2@example.com",
			Username:    baseUser.Username, // Duplicate username
			PhoneNumber: "1000000002",
			Password:    "TestPassword123!",
		}
		userJSON, _ := json.Marshal(duplicateUser)
		req, _ := http.NewRequest("POST", "/auth/signup", bytes.NewBuffer(userJSON))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status 400 for duplicate username, got %v", status)
		}
		var response map[string]string
		json.Unmarshal(rr.Body.Bytes(), &response)
		if response["error"] != repository.ErrorUsernameExists.Error() {
			t.Errorf("Expected error '%s', got '%s'", repository.ErrorUsernameExists.Error(), response["error"])
		}
	})

	t.Run("Duplicate Phone Number", func(t *testing.T) {
		duplicateUser := model.UserSignUp{
			Email:       "new_email_3@example.com",
			Username:    "new_username_3",
			PhoneNumber: baseUser.PhoneNumber, // Duplicate phone number
			Password:    "TestPassword123!",
		}
		userJSON, _ := json.Marshal(duplicateUser)
		req, _ := http.NewRequest("POST", "/auth/signup", bytes.NewBuffer(userJSON))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Expected status 400 for duplicate phone number, got %v", status)
		}
		var response map[string]string
		json.Unmarshal(rr.Body.Bytes(), &response)
		if response["error"] != repository.ErrorPhoneNumberExists.Error() {
			t.Errorf("Expected error '%s', got '%s'", repository.ErrorPhoneNumberExists.Error(), response["error"])
		}
	})
}
