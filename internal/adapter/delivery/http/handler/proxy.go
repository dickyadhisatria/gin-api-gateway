package handler

import (
	"errors"
	"gin-api-gateway/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
)

// ProxyHandler memegang dependensi untuk menangani request perutean HTTP.
// ProxyHandler holds dependencies to handle request routing HTTP handlers.
type ProxyHandler struct {
	proxyUseCase usecase.ProxyUseCase
}

// NewProxyHandler menginisialisasi controller handler baru.
// NewProxyHandler initializes a new handler controller instance.
//
// Parameters:
//   - uc (usecase.ProxyUseCase): Kasus penggunaan perutean proxy / Proxy routing UseCase dependency.
//
// Returns:
//   - *ProxyHandler: Controller yang terisi dependensi UseCase / Controller loaded with UseCase dependency.
func NewProxyHandler(uc usecase.ProxyUseCase) *ProxyHandler {
	return &ProxyHandler{proxyUseCase: uc}
}

// Redirect mengembalikan gin.HandlerFunc yang mengalihkan request ke url target.
// Redirect returns a gin.HandlerFunc that routes incoming request to the target url.
//
// Parameters:
//   - targetURL (string): URL dari microservice hilir tujuan / Downstream microservice target url.
//
// Returns:
//   - gin.HandlerFunc: Handler rute Gin / Gin route handler middleware.
func (h *ProxyHandler) Redirect(targetURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Mengeksekusi pengalihan permintaan ke target microservice melalui ProxyUseCase
		// 1. Execute forwarding request to target microservice using ProxyUseCase
		err := h.proxyUseCase.Execute(c.Writer, c.Request, targetURL)
		if err != nil {
			// 2. Memeriksa apakah error disebabkan oleh Circuit Breaker yang sedang terbuka (Open/Half-Open state)
			// 2. Evaluate if error was triggered by a tripped (Open/Half-Open state) Circuit Breaker
			if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "Server tujuan sedang mengalami gangguan teknis. Silakan coba beberapa saat lagi.",
				})
				c.Abort()
				return
			}

			// 3. Jika terjadi kesalahan koneksi dan header respon belum dikirim, berikan respon error 500
			// 3. If standard network failure occurs and no headers were written yet, write 500 Internal Error
			if !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to target service"})
				c.Abort()
			}
		}
	}
}
