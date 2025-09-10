// File: backend/internal/api/handler/auth_handler.go (CORRECTED for alexedwards/argon2id)
package handler

import (
	"backend/internal/database"
	"backend/internal/model"
	"backend/internal/service"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/argon2id" // <--- CHANGED IMPORT HERE
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	// Removed golang.org/x/crypto/argon2 as it's no longer directly used here
)

// SignUpHandler godoc
//
//	@Summary		User Sign Up
//	@Description	Creates a new user account after validating username, email, and password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			userDetails	body		model.UserSignUp	true	"User sign up details"
//	@Success		201			{string}	string				"Account created successfully"
//	@Failure		400			{string}	string				"Validation error or bad request"
//	@Failure		500			{string}	string				"Internal server error"
//	@Router			/auth/signup [post]
func SignUpHandler(c *gin.Context) {
	var userDetails model.UserSignUp
	if err := c.ShouldBindJSON(&userDetails); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.SignUpService(c, userDetails); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"result": "account created successfully"})
}

func findUserByEmail(email string) (*model.User, error) {
	user := &model.User{}

	if dbService == nil {
		slog.Error("Database service not set in auth handler.")
		return nil, gin.Error{Err: nil, Type: gin.ErrorTypePrivate, Meta: "Database service not initialized"}
	}

	rawDB := dbService.GetRawDB()
	if rawDB == nil {
		slog.Error("Raw database connection is nil in auth handler.")
		return nil, gin.Error{Err: nil, Type: gin.ErrorTypePrivate, Meta: "Raw database connection unavailable"}
	}

	query := "SELECT id, email, username, hashed_password FROM users WHERE email = $1"

	row := rawDB.QueryRow(query, email)

	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.HashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, gin.Error{Err: err, Type: gin.ErrorTypePrivate, Meta: "User not found"}
		}
		slog.Error("Database query error in findUserByEmail", slog.Any("error", err))
		return nil, gin.Error{Err: err, Type: gin.ErrorTypePrivate, Meta: "Database query error"}
	}

	return user, nil
}

func GenerateJWT(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(JwtSecretKey)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func LoginHandler(c *gin.Context) {
	var loginDetails struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&loginDetails); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	user, err := findUserByEmail(loginDetails.Email)
	if err != nil {
		slog.Warn("Login failed for email (user not found/db error)", slog.String("email", loginDetails.Email), slog.Any("error", err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// --- CORRECTED ARGON2 PASSWORD VERIFICATION using alexedwards/argon2id ---
	// user.HashedPassword must contain the *full* Argon2 hash string, including parameters and salt.
	match, err := argon2id.ComparePasswordAndHash(loginDetails.Password, user.HashedPassword)
	if err != nil {
		// Handle error (e.g., hash format is invalid, unexpected issue)
		slog.Error("Error comparing password and hash", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error during password verification"})
		return
	}
	if !match {
		slog.Warn("Login failed for email (password mismatch)", slog.String("email", loginDetails.Email))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	// --- END CORRECTION ---

	token, err := GenerateJWT(user.ID)
	if err != nil {
		slog.Error("Error generating token", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login successful!", "token": token})
}