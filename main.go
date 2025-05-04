package main

import (
	"log"

	auth "github.com/Blockchain-Payment-Infrastructure/backend/auth/handlers"
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

	r := SetupRouter()
	if err := r.Run(); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
