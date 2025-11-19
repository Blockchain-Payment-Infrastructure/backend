package server

import (
	"backend/internal/api/handler"
	"backend/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(middleware.StructuredLogger())

	r.GET("/", s.HelloWorldHandler)
	r.GET("/health", s.healthHandler)

	auth := r.Group("/auth")
	{
		auth.POST("/signup", handler.SignUpHandler)
		auth.POST("/login", handler.LoginHandler)
		auth.POST("/refresh", handler.RefreshTokenHandler)
		auth.POST("/logout", handler.LogoutHandler)
	}

	wallet := r.Group("/wallet")

	wallet.Use(middleware.AuthMiddleware())
	{
		wallet.GET("/addresses/:phone_number", handler.WalletAddressFromPhoneHandler)
		wallet.POST("/connect", handler.ConnectWalletHandler)
		wallet.GET("/balance/:address", handler.GetWalletBalanceHandler)
		wallet.GET("/balances", handler.GetUserWalletBalancesHandler)
	}

	payments := r.Group("/payments")
	payments.Use(middleware.AuthMiddleware())
	{
		payments.POST("", handler.CreatePaymentHandler)
		payments.GET("", handler.GetUserPaymentsHandler)
		payments.GET("/stats", handler.GetPaymentStatsHandler)
		payments.GET("/:id", handler.GetPaymentHandler)
		payments.POST("/:id/refresh", handler.RefreshPaymentStatusHandler)
		payments.GET("/tx/:hash", handler.GetPaymentByTransactionHashHandler)
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
