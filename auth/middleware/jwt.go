package middleware

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Blockchain-Payment-Infrastructure/backend/auth/services"
	"github.com/Blockchain-Payment-Infrastructure/backend/auth/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthenticateJWTMiddleware(c *gin.Context) {
	tokenString := c.GetHeader("BEARER")
	if tokenString == "" {
		log.Println("Token missing in header")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Method)
		}

		return utils.JWTSecret, nil
	})
	if err != nil {
		log.Printf("User token validation failed: %v", err)
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if time.Now().Unix() > int64(claims["exp"].(float64)) {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		user, err := services.GetUserByID(claims["sub"].(string))
		if err != nil {
			log.Printf("User fetching failed: %v", err)
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		c.Set("user", user)
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	c.Next()
}
