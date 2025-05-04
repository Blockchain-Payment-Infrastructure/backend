package router

import (
	"github.com/Blockchain-Payment-Infrastructure/backend/auth/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {
	auth := r.Group("/api/auth")

	auth.POST("/register", handlers.RegisterHandler)
	auth.POST("/login", handlers.LoginHandler)
}
