// File: backend/internal/server/routes.go (MODIFIED: Integrated account settings routes, CORS)
package server

import (
	"backend/internal/api/handler"
	"backend/internal/middleware"
	"net/http"

<<<<<<< HEAD
	"github.com/gin-contrib/cors" // <--- ADD THIS IMPORT for CORS configuration
=======
	// <--- ADD THIS IMPORT for CORS configuration
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.New()

	// Apply global middleware
	r.Use(gin.Recovery())
	r.Use(middleware.StructuredLogger())
	// CORS Middleware - configure as per your frontend needs
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://yourfrontend.com"}, // IMPORTANT: Adjust this to your frontend's actual URL(s)
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	// Database service is available on the server as `s.db` and
	// repository/service packages can obtain DB connections via their own
	// helpers (e.g., calling `database.New`) — avoid global setters here
	// to keep initialization explicit and clear.

	// --- Public Routes ---
	r.GET("/", s.HelloWorldHandler)
	r.GET("/health", s.healthHandler)

	auth := r.Group("/auth")
	{
		auth.POST("/signup", handler.SignUpHandler)
		auth.POST("/login", handler.LoginHandler)
		auth.POST("/refresh", handler.RefreshTokenHandler) // Requires handler.RefreshTokenHandler
		auth.POST("/logout", handler.LogoutHandler)        // Requires handler.LogoutHandler
	}

	// --- Protected Routes (require AuthMiddleware) ---
	// Create a protected group that applies the AuthMiddleware. Using
	// root-level paths (e.g. /account) preserves existing client routes.
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())

	// Wallet Features
	wallet := protected.Group("/wallet")
	{
		wallet.GET("/addresses/:phone_number", handler.WalletAddressFromPhoneHandler)
		wallet.POST("/connect", handler.ConnectWalletHandler)
		wallet.GET("/balance/:address", handler.GetWalletBalanceHandler)
		wallet.GET("/balances", handler.GetUserWalletBalancesHandler)
	}

	// Payments Features
	payments := protected.Group("/payments")
	{
		payments.POST("", handler.CreatePaymentHandler)
		payments.GET("", handler.GetUserPaymentsHandler)
		payments.GET("/stats", handler.GetPaymentStatsHandler)
		payments.GET("/:id", handler.GetPaymentHandler)
		payments.POST("/:id/refresh", handler.RefreshPaymentStatusHandler)
		payments.GET("/tx/:hash", handler.GetPaymentByTransactionHashHandler)
	}

	// --- Account Settings Features (NEW) ---
	account := protected.Group("/account") // Group related account settings routes
	{
		account.PATCH("/change-password", handler.ChangePasswordHandler)
		account.PATCH("/update-email", handler.UpdateEmailHandler)
		account.DELETE("/delete", handler.DeleteAccountHandler)
	}

<<<<<<< HEAD
	// Example Protected Dashboard route (you might already have this or remove it)
	protected.GET("/dashboard", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome to your secure dashboard!"})
	})

=======
>>>>>>> a7fcdf6fcb199bb557aabcd039480382d05b095d
	return r
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := map[string]string{"message": "Hello World"}
	c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
