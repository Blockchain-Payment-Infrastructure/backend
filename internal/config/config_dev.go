//go:build !prod

package config

import "github.com/gin-gonic/gin"

const (
	SecureCookie = true
	AppMode      = gin.DebugMode
)
