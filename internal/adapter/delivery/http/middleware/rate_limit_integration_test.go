package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gin-api-gateway/internal/adapter/delivery/http/middleware"
	"gin-api-gateway/internal/adapter/ratelimit"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func TestRateLimiter_Integration_Redis(t *testing.T) {
	// Connect to local Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()

	// Ensure Redis is reachable
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis is not running on localhost:6379, skipping integration test:", err)
	}

	// Clean up previous rate limit keys for testing IP
	testIP := "192.168.1.100"
	keys, _ := rdb.Keys(ctx, "rate_limit:*").Result()
	for _, key := range keys {
		rdb.Del(ctx, key)
	}

	// Initialize RedisRateLimiter with a low limit: 5 requests per minute
	limiter, err := ratelimit.NewRedisRateLimiter(rdb, 5, time.Minute)
	if err != nil {
		t.Fatalf("failed to create redis rate limiter: %v", err)
	}

	// Setup Gin
	gin.SetMode(gin.TestMode)
	r := gin.New()
	
	// Mock IP extraction
	r.Use(func(c *gin.Context) {
		c.Request.RemoteAddr = testIP + ":1234"
		c.Next()
	})
	r.Use(middleware.RateLimiter(limiter))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Perform 5 successful requests
	for i := 1; i <= 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Request %d expected 200, got %d", i, w.Code)
		}
	}

	// The 6th request should fail with 429 Too Many Requests
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("6th Request expected 429, got %d. Response: %s", w.Code, w.Body.String())
	}

	// Clean up
	keys, _ = rdb.Keys(ctx, "rate_limit:*").Result()
	for _, key := range keys {
		rdb.Del(ctx, key)
	}
}
