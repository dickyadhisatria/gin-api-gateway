package main

import (
	"context"
	"gin-api-gateway/internal/adapter/auth"
	"gin-api-gateway/internal/adapter/delivery/http/handler"
	"gin-api-gateway/internal/adapter/delivery/http/middleware"
	"gin-api-gateway/internal/adapter/proxy"
	"gin-api-gateway/internal/adapter/ratelimit"
	"gin-api-gateway/internal/domain/service"
	"gin-api-gateway/internal/infrastructure/config"
	"gin-api-gateway/internal/usecase"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// main adalah entrypoint utama aplikasi yang bertindak sebagai Composition Root.
// main is the primary entrypoint acting as the Composition Root of the application.
func main() {
	// 1. Memuat konfigurasi aplikasi dari variabel lingkungan (.env)
	// 1. Load application configuration from environment variables (.env)
	cfg := config.LoadConfig()

	// 2. Mengatur mode jalannya framework Gin ke ReleaseMode
	// 2. Set the Gin framework mode to ReleaseMode
	gin.SetMode(gin.ReleaseMode)

	// 3. Menginisialisasi adapter untuk Reverse Proxy dan JWT Authenticator
	// 3. Initialize adapters for Reverse Proxy and JWT Authenticator
	proxyClient := proxy.NewHTTPProxy()
	authenticator := auth.NewJWTAuthenticator(cfg.JwtSecret)

	// 3a. Menginisialisasi Rate Limiter (Redis jika dikonfigurasi dan aktif, jika tidak gunakan Memori)
	// 3a. Initialize Rate Limiter (Redis if configured and active, otherwise fall back to Memory)
	var rateLimiter service.RateLimiter
	if cfg.RedisURL != "" {
		rdb := redis.NewClient(&redis.Options{
			Addr:     cfg.RedisURL,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})
		
		// Verifikasi apakah koneksi ke Redis aktif sebelum menggunakannya
		// Verify if connection to Redis is active before using it
		pingCtx, pingCancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := rdb.Ping(pingCtx).Err()
		pingCancel()

		if err != nil {
			slog.Warn("Redis is down or unreachable, falling back to memory rate limiter", "error", err)
			rateLimiter = ratelimit.NewMemoryRateLimiter(60, time.Minute)
		} else {
			var initErr error
			rateLimiter, initErr = ratelimit.NewRedisRateLimiter(rdb, 60, time.Minute)
			if initErr != nil {
				slog.Warn("Failed to initialize Redis rate limiter, falling back to memory store", "error", initErr)
				rateLimiter = ratelimit.NewMemoryRateLimiter(60, time.Minute)
			} else {
				slog.Info("Redis rate limiter initialized successfully")
			}
		}
	} else {
		rateLimiter = ratelimit.NewMemoryRateLimiter(60, time.Minute)
		slog.Info("Memory rate limiter initialized successfully (No REDIS_URL configured)")
	}

	// 4. Menginisialisasi usecase untuk perutean proxy (ProxyUseCase)
	// 4. Initialize the UseCase for proxy routing (ProxyUseCase)
	proxyUseCase := usecase.NewProxyUseCase(proxyClient)

	// 5. Menginisialisasi controller/handler untuk menangani pengalihan HTTP
	// 5. Initialize the controller/handler for handling HTTP redirection
	proxyHandler := handler.NewProxyHandler(proxyUseCase)

	// 6. Membuat instance engine Gin baru yang bersih tanpa middleware bawaan
	// 6. Construct a fresh Gin engine instance without default middlewares
	r := gin.New()

	// 7. Memasang middleware pencatat log (logger) dan pembatas laju (rate limiter) secara global
	// 7. Attach structured logger and rate limiter middlewares globally
	r.Use(middleware.JSONLogger())
	r.Use(middleware.RateLimiter(rateLimiter))

	// 8. Mendaftarkan rute publik untuk proses autentikasi (tidak memerlukan login)
	// 8. Register public routes for authentication (no login required)
	r.POST("/api/auth/*proxyPath", proxyHandler.Redirect(cfg.AuthService))

	// 9. Mendaftarkan rute privat dengan pengamanan token JWT melalui grup router
	// 9. Register private routes secured with JWT token authentication via a router group
	api := r.Group("/api/v1")
	api.Use(middleware.ValidateJWT(authenticator))
	{
		// 10. Mengarahkan rute privat ke layanan mikro terkait (User, Product, Order)
		// 10. Forward private routes to respective microservices (User, Product, Order)
		api.Any("/user/*proxyPath", proxyHandler.Redirect(cfg.UserService))
		api.Any("/product/*proxyPath", proxyHandler.Redirect(cfg.ProductService))
		api.Any("/order/*proxyPath", proxyHandler.Redirect(cfg.OrderService))
	}

	// 11. Menampilkan log informasi bahwa server gateway telah berjalan
	// 11. Print log information indicating the gateway server is active
	slog.Info("Gateway is running on port", "port", cfg.Port)

	// 12. Membuat konfigurasi server HTTP dengan timeout baca dan tulis
	// 12. Build HTTP server configuration with read and write timeouts
	srv := &http.Server{
		Addr:         cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// 13. Menjalankan server HTTP di dalam goroutine agar tidak memblokir proses shutdown
	// 13. Execute the HTTP server inside a goroutine to prevent blocking the shutdown sequence
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", "error", err)
		}
	}()

	// 14. Menunggu sinyal terminasi OS (SIGINT atau SIGTERM) untuk memulai shutdown anggun
	// 14. Wait for OS termination signals (SIGINT or SIGTERM) to initiate graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down gateway gracefully...")

	// 15. Memberikan waktu tenggang 10 detik untuk menyelesaikan request yang sedang berjalan
	// 15. Provide a 10-second grace period to drain active HTTP requests
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Gateway forced to shutdown", "error", err)
	}

	slog.Info("Gateway exited cleanly")
}
