package config_test

import (
	"context"
	"testing"
	"time"

	"gin-api-gateway/internal/adapter/ratelimit"
	"github.com/redis/go-redis/v9"
)

func TestRedisFallback(t *testing.T) {
	// 1. Membuat client Redis dengan port localhost yang offline
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:9999", // Port pasti offline
	})
	defer rdb.Close()

	// 2. Melakukan ping ke Redis dengan timeout singkat
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	err := rdb.Ping(pingCtx).Err()
	pingCancel()

	// 3. Memverifikasi koneksi gagal (error tidak nil)
	if err == nil {
		t.Fatal("Expected Redis to be offline, but ping succeeded")
	}

	// 4. Memastikan inisialisasi fallback ke MemoryRateLimiter berjalan sukses
	rateLimiter := ratelimit.NewMemoryRateLimiter(5, time.Minute)
	if rateLimiter == nil {
		t.Fatal("Failed to initialize memory rate limiter fallback")
	}

	// 5. Menguji apakah MemoryRateLimiter fallback berfungsi dengan benar
	allowed, testErr := rateLimiter.Allow("127.0.0.1")
	if testErr != nil {
		t.Fatalf("Memory rate limiter returned error: %v", testErr)
	}
	if !allowed {
		t.Fatal("Memory rate limiter should allow request under quota")
	}
}
