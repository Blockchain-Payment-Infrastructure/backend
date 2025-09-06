package tests

import (
	"backend/internal/api/handler"
	"backend/internal/database"
	"backend/internal/model"
	"backend/internal/repository"
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUserSignup(t *testing.T) {
	r := gin.Default()

	teardown, dsn, err := MustStartPostgresContainer()
	if err != nil {
		slog.Error("could not start postgres container:", slog.Any("error", err))
	}

	database.New(dsn)
	database.Migrate("file://../db/migrations")

	auth := r.Group("/auth")
	auth.POST("/signup", handler.SignUpHandler)

	user := model.UserSignUp{
		Email:    "testing@abcd.com",
		Username: "testing",
		Password: "H*mUhZ655mJo$$@K",
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/auth/signup", bytes.NewBuffer(userJSON))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	expected := "{\"result\":\"account created successfully\"}"
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	ctx := context.Background()
	if exists, err := repository.UserExists(ctx, user.Email); err != nil && !exists {
		t.Fatal("User not found in database")
	}

	Cleanup(teardown)
}
