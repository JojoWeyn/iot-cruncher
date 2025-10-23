package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		// TODO: identity
		c.Set("user_id", "a7745bd5-a8ab-40a6-b776-a802ff75f9d9")
		c.Next()
	}
}
