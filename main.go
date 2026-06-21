package main

import (
	"gin-api-gateway/config"
	"gin-api-gateway/middleware"
	"gin-api-gateway/proxy"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Ambil data dari file .env
	cfg := config.LoadConfig()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(middleware.RateLimiter())

	// 2. Buat Sekring (Circuit Breaker) khusus untuk masing-masing Layanan
	cbUser := proxy.CreateCircuitBreaker("User-Service")
	cbOrder := proxy.CreateCircuitBreaker("Order-Service")

	// 3. Pasang keamanan API Key dengan data rahasia dari .env
	api := r.Group("/api")
	api.Use(middleware.CekApiKey(cfg.ApiKey))
	{
		// Akses User Service (Dilindungi Sekring cbUser)
		api.Any("/users/*proxyPath", func(c *gin.Context) {
			proxy.RedirectWithCircuitBreaker(c, cfg.UserService, cbUser)
		})

		// Akses Order Service (Dilindungi Sekring cbOrder)
		api.Any("/orders/*proxyPath", func(c *gin.Context) {
			proxy.RedirectWithCircuitBreaker(c, cfg.OrderService, cbOrder)
		})
	}

	r.Run(cfg.Port)
}
