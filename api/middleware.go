package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func apiKeyAuth(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if apiKey == "" {
			c.Next()
			return
		}
		if c.GetHeader("X-Api-Key") != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid or missing API key"})
			return
		}
		c.Next()
	}
}
