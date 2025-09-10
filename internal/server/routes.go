// File: backend/internal/server/routes.go (MODIFIED to inject DB into service layer)
package server

import (
	"backend/internal/api/handler"
	"backend/internal/middleware"
	"backend/internal/service" // <--- Import your service package
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.GET("/", s.HelloWorldHandler)
	r.GET("/health", s.healthHandler)

	// --- Inject DB service into handler and service layers ---
	handler.SetDBService(s.db)
	service.SetDBService(s.db) // <--- NEW: Inject DB service into your service package

	auth := r.Group("/auth")
	{
		auth.POST("/signup", handler.SignUpHandler)
		auth.POST("/login", handler.LoginHandler)
	}

	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/dashboard", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Welcome to your secure dashboard!"})
		})
	}

	return r
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := map[string]string{"message": "Hello World"}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
