package main

import (
	"log"
	"net/http"

	auth "github.com/Blockchain-Payment-Infrastructure/backend/auth/handlers"
	jwt_middleware "github.com/Blockchain-Payment-Infrastructure/backend/auth/middleware"
	auth_jwt "github.com/Blockchain-Payment-Infrastructure/backend/auth/utils"

	"github.com/Blockchain-Payment-Infrastructure/backend/config"
	"github.com/Blockchain-Payment-Infrastructure/backend/db"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Authentication
	authentication := r.Group("/api/auth")
	{
		authentication.POST("/register", auth.RegisterHandler)
		authentication.POST("/login", auth.LoginHandler)
		authentication.GET("/validate", jwt_middleware.AuthenticateJWTMiddleware, func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "valid token"})
		})
	}

	return r
}

func main() {
	if err := config.LoadEnv(); err != nil {
		return
	}

	auth_jwt.InitialiseJWT(config.JWTSecret)

	err := db.InitDB(config.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	db.RunMigrations()
	db.DB.SetMaxOpenConns(20)

	r := SetupRouter()
	if err := r.Run(); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
