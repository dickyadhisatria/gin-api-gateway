package middleware

import (
	"gin-api-gateway/internal/domain/service"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RateLimiter membatasi frekuensi request berdasarkan IP client secara global/terdistribusi.
// RateLimiter restrains request frequency based on client IP globally or distributively.
//
// Parameters:
//   - limiter (service.RateLimiter): Implementasi pembatas laju domain / Domain rate limiter implementation.
//
// Returns:
//   - gin.HandlerFunc: Middleware rute Gin / Gin middleware handler function.
func RateLimiter(limiter service.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Mendapatkan alamat IP Client dari context request
		// 1. Retrieve the client IP address from request context
		ip := c.ClientIP()
		
		// 2. Mengecek apakah IP Client tersebut masih berada di bawah limit kuota yang ditentukan
		// 2. Evaluate if the client IP is still under specified request threshold
		allowed, err := limiter.Allow(ip)
		if err != nil {
			// Fail-open: log error dan izinkan request lewat agar gangguan pada rate limiter backend tidak memicu pemadaman massal (outage)
			// Fail-open: log error and allow request to bypass so rate limiter backend failures don't cause cascading system outages
			slog.Error("Rate limiter check failed, failing open", "error", err, "client_ip", ip)
			c.Next()
			return
		}

		// 3. Jika IP Client melampaui limit, gagalkan request dengan respon status 429 Too Many Requests
		// 3. If IP exceeds limit, fail request and return 429 Too Many Requests HTTP status
		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Please try again later."})
			c.Abort()
			return
		}

		// 4. Melanjutkan eksekusi ke handler berikutnya apabila diizinkan
		// 4. Resume route processing to next handler if allowed
		c.Next()
	}
}
