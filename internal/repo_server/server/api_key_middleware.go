package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type ApiKeyAuthMiddlewareConfig struct {
	AllowedApiKeys    []string
	UnprotectedRoutes []string
}

func GetApiKeyAuthMiddleware(config *ApiKeyAuthMiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {

		for _, route := range config.UnprotectedRoutes {
			if c.Request.URL.Path == route {
				c.Next()
				return
			}
		}

		apiKeyHeaderValue := c.GetHeader("X-API-Key")
		if apiKeyHeaderValue == "" {
			c.AbortWithError(401, fmt.Errorf("no api key provided"))
		}

		for _, validApiKey := range config.AllowedApiKeys {
			if apiKeyHeaderValue == validApiKey {
				c.Next()
				return
			}
		}
		c.AbortWithError(401, fmt.Errorf("invalid API key"))
	}
}
