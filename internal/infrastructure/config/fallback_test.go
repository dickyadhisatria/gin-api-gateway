package config_test

import (
	"testing"
	"time"

	"gin-api-gateway/internal/adapter/ratelimit"
)

func TestMemoryRateLimiterInitialization(t *testing.T) {
	// 1. Memastikan inisialisasi MemoryRateLimiter berjalan sukses
	rateLimiter := ratelimit.NewMemoryRateLimiter(5, time.Minute)
	if rateLimiter == nil {
		t.Fatal("Failed to initialize memory rate limiter")
	}

	// 2. Menguji apakah MemoryRateLimiter berfungsi dengan benar
	allowed, testErr := rateLimiter.Allow("127.0.0.1")
	if testErr != nil {
		t.Fatalf("Memory rate limiter returned error: %v", testErr)
	}
	if !allowed {
		t.Fatal("Memory rate limiter should allow request under quota")
	}
}
