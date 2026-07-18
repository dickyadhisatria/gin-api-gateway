package middleware

import (
	"gin-api-gateway/internal/domain/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ValidateJWT adalah middleware keamanan untuk memvalidasi token JWT dan otorisasi.
// ValidateJWT is a security middleware that validates incoming JWT tokens and authorizations.
//
// Parameters:
//   - auth (service.Authenticator): Implementasi validator token domain / Domain token validator implementation.
//
// Returns:
//   - gin.HandlerFunc: Middleware rute Gin / Gin middleware handler function.
func ValidateJWT(auth service.Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Membaca header "Authorization" dari permintaan HTTP masuk
		// 1. Read the "Authorization" header from incoming HTTP request
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not found! Please login first."})
			c.Abort()
			return
		}

		// 2. Memisahkan tipe token dan string token (format harus "Bearer <token>")
		// 2. Separate token type from token string content (format must be "Bearer <token>")
		payload := strings.Split(authHeader, " ")
		if len(payload) != 2 || payload[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token format is incorrect (Must use Bearer Token)"})
			c.Abort()
			return
		}

		tokenStr := payload[1]

		// 3. Memverifikasi keabsahan token JWT menggunakan domain Authenticator service
		// 3. Verify JWT token validity using the domain Authenticator service
		claims, err := auth.Verify(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is invalid or has expired!"})
			c.Abort()
			return
		}

		// 4. Menyuntikkan user_id dan role hasil ekstrak ke header HTTP internal untuk diteruskan ke downstream microservices
		// 4. Inject extracted user_id and role into internal HTTP headers to forward to downstream microservices
		c.Request.Header.Set("X-User-ID", claims.UserID)
		c.Request.Header.Set("X-User-Role", claims.Role)

		// 5. Melanjutkan eksekusi permintaan ke handler berikutnya
		// 5. Proceed request execution to the next handler/middleware
		c.Next()
	}
}
