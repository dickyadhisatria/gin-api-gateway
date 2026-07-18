package middleware

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// generateRequestID memproduksi UUID v4 acak string untuk ID pelacakan request.
// generateRequestID produces a cryptographically secure random request ID for tracing.
//
// Returns:
//   - string: Token ID acak 32 karakter heksadesimal / Hexadecimal request trace ID token.
func generateRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "gen-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return fmt.Sprintf("%x", b)
}

// JSONLogger mencatat riwayat lalu lintas request masuk dan durasi respon dalam format JSON terstruktur.
// JSONLogger writes request traffic details and response latency in structured JSON logs.
//
// Returns:
//   - gin.HandlerFunc: Middleware rute Gin / Gin middleware handler function.
func JSONLogger() gin.HandlerFunc {
	// 1. Menginisialisasi slog JSON handler baru yang menulis log terstruktur ke stdout
	// 1. Initialize a new slog JSON handler writing structured logs to stdout
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return func(c *gin.Context) {
		// 2. Mencatat waktu mulai eksekusi permintaan
		// 2. Record request execution start timestamp
		start := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path

		// 3. Mengambil Request ID dari header atau membuatnya secara acak jika belum ada
		// 3. Get Request ID from incoming header or generate a random one if empty
		requestID := c.Request.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			c.Request.Header.Set("X-Request-ID", requestID)
		}

		// 4. Mengambil IP Client atau menetapkan sebagai "anonymous" jika tidak ditemukan
		// 4. Read client IP or set fallback value if missing
		clientIP := c.ClientIP()
		if clientIP == "" {
			clientIP = "anonymous"
		}

		// 5. Meneruskan eksekusi ke handler berikutnya
		// 5. Forward request processing to next middleware or route handler
		c.Next()

		// 6. Mengambil status kode HTTP respon setelah eksekusi selesai
		// 6. Fetch HTTP status code after execution completes
		status := c.Writer.Status()
		
		// 7. Menghitung latensi eksekusi dalam satuan milidetik
		// 7. Calculate elapsed response latency in milliseconds
		duration := time.Since(start)
		latency := duration.Seconds() * 1000 // ms

		// 8. Mencatat seluruh data pemanggilan HTTP tersebut ke log dalam format JSON terstruktur
		// 8. Write the request details as a structured JSON record
		log.Info("Request completed",
			slog.String("request_id", requestID),
			slog.String("client_ip", clientIP),
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Float64("latency_ms", latency),
		)
	}
}
