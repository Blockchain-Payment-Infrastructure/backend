package tests

import (
	"backend/internal/api/handler"
	"backend/internal/database"
	"backend/internal/model"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUserLogin(t *testing.T) {
	database.New(testDSN)
	database.Migrate("file://../db/migrations")

	r := gin.Default()
	auth := r.Group("/auth")
	auth.POST("/signup", handler.SignUpHandler)
	auth.POST("/login", handler.LoginHandler)

	user := model.UserSignUp{
		Email:    "testing@abcde.com",
		Username: "testings",
		Password: "H*mUhZ655mJo$$@Ka",
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

	userLogin := model.UserLogin{
		Email:    "testing@abcde.com",
		Password: "H*mUhZ655mJo$$@Ka",
	}

	userJSON, err = json.Marshal(userLogin)
	if err != nil {
		t.Fatal(err)
	}

	req, err = http.NewRequest("POST", "/auth/login", bytes.NewBuffer(userJSON))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application.json")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "Login successful!"
	m := map[string]string{}
	json.Unmarshal(rr.Body.Bytes(), &m)
	if m["message"] != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", m["message"], expected)
	}

}
