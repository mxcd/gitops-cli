package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type ApiKeyAuthMiddlewareConfig struct {
	AllowedApiKeys []string
}

func GetApiKeyAuthMiddleware(config *ApiKeyAuthMiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKeyHeaderValue := c.GetHeader("X-API-Key")
		if apiKeyHeaderValue == "" {
			c.AbortWithError(401, fmt.Errorf("no api key provided"))
		}

		for _, validApiKey := range config.AllowedApiKeys {
			if apiKeyHeaderValue == validApiKey {
				c.Next()
			}
		}
		c.AbortWithError(401, fmt.Errorf("invalid API key"))
	}
}
