package proxy

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
)

// Buat pengaturan Circuit Breaker global
func CreateCircuitBreaker(serviceName string) *gobreaker.CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        serviceName,
		MaxRequests: 3,                // Uji coba kirim 3 data saat status Half-Open
		Interval:    5 * time.Second,  // Reset hitungan error setiap 5 detik
		Timeout:     10 * time.Second, // Durasi sekring mati (status Open) sebelum dicoba lagi
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Jika terjadi 3 kali error beruntun, sekring langsung PUTUS (Open)
			return counts.ConsecutiveFailures >= 3
		},
	}
	return gobreaker.NewCircuitBreaker(settings)
}

// RedirectWithCircuitBreaker melempar data dengan pengaman sekring otomatis
func RedirectWithCircuitBreaker(c *gin.Context, targetURL string, cb *gobreaker.CircuitBreaker) {
	remote, err := url.Parse(targetURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Alamat server tujuan salah"})
		c.Abort()
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	// Jalankan proxy di dalam pengawasan Circuit Breaker
	_, errCB := cb.Execute(func() (interface{}, error) {

		// Buat sebuah penangkap status error HTTP khusus dari server belakang
		penangkapStatus := &statusResponseWritter{ResponseWriter: c.Writer, status: http.StatusOK}

		proxy.ServeHTTP(penangkapStatus, c.Request)

		// Jika server belakang mengembalikan status rusak (500 keatas), hitung sebagai error
		if penangkapStatus.status >= 500 {
			return nil, errors.New("server belakang error")
		}

		return nil, nil
	})

	// Jika sekring sedang PUTUS, langsung tolak pengguna di sini
	if errCB != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Server tujuan sedang mengalami gangguan teknis. Silakan coba beberapa saat lagi.",
		})
		c.Abort()
		return
	}
}

// Struktur pembantu untuk membaca status kode HTTP (seperti 200, 404, 500)
type statusResponseWritter struct {
	gin.ResponseWriter
	status int
}

func (w *statusResponseWritter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
