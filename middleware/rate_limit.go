package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// RateLimiter membatasi pengguna maksimal hanya 60 kali akses per menit
func RateLimiter() gin.HandlerFunc {
	// Tentukan kuota: 60 kali per menit
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  60,
	}

	// Simpan data hit di dalam memori internal komputer
	store := memory.NewStore()
	instance := limiter.New(store, rate)

	return func(c *gin.Context) {
		// Batasi berdasarkan alamat IP komputer pengguna
		ip := c.ClientIP()
		context, err := instance.Get(c, ip)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengecek limit"})
			c.Abort()
			return
		}

		// Jika akses melebihi batas, kunci pintu masuk
		if context.Reached {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Akses terlalu cepat! Silakan tunggu semenit lagi."})
			c.Abort()
			return
		}
		c.Next()
	}
}
