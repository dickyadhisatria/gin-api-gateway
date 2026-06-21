package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CekApiKey memeriksa token berdasarkan data dari file .env
func CekApiKey(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		secretkey := c.GetHeader("X-API-KEY")

		if secretkey != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak! API Key salah atau tidak ada."})
			c.Abort()
			return
		}
		c.Next()
	}
}
