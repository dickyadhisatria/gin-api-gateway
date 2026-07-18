package ratelimit

import (
	"context"
	"gin-api-gateway/internal/domain/service"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	limiterredis "github.com/ulule/limiter/v3/drivers/store/redis"
)

type redisRateLimiter struct {
	limiterInstance *limiter.Limiter
}

// NewRedisRateLimiter menginisialisasi instansi pembatas laju terdistribusi berbasis Redis.
// NewRedisRateLimiter initializes a distributed Redis-based rate limiter instance.
//
// Parameters:
//   - client (redis.UniversalClient): Client Redis yang aktif / Active Redis client instance.
//   - limit (int64): Jumlah maksimum permintaan yang diizinkan / Maximum request count allowed.
//   - period (time.Duration): Periode waktu perputaran limit / Limit reset time period.
//
// Returns:
//   - service.RateLimiter: Implementasi antarmuka RateLimiter / RateLimiter interface implementation.
//   - error: Error inisialisasi store Redis / Redis store initialization error.
func NewRedisRateLimiter(client redis.UniversalClient, limit int64, period time.Duration) (service.RateLimiter, error) {
	rate := limiter.Rate{
		Period: period,
		Limit:  limit,
	}

	// 1. Menginisialisasi store Redis untuk ulule/limiter dengan client redis.UniversalClient
	// 1. Initialize the Redis store for ulule/limiter using the redis.UniversalClient
	store, err := limiterredis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix: "rate_limit:",
	})
	if err != nil {
		return nil, err
	}

	// 2. Membuat instance limiter baru dengan store Redis
	// 2. Build the new limiter instance connected to the Redis store
	instance := limiter.New(store, rate)

	return &redisRateLimiter{
		limiterInstance: instance,
	}, nil
}

// Allow mengecek apakah IP klien diperbolehkan mengirimkan request berdasarkan kuota Redis global.
// Allow checks if the client IP is allowed to send request based on the global Redis quota.
//
// Parameters:
//   - ip (string): Alamat IP klien / Client IP address.
//
// Returns:
//   - bool: True jika diizinkan, False jika dibatasi / True if allowed, False if limited.
//   - error: Error koneksi atau pembacaan Redis / Redis query or connection error.
func (r *redisRateLimiter) Allow(ip string) (bool, error) {
	// 1. Mengecek limit kuota secara terdistribusi di Redis untuk IP Client tertentu
	// 1. Evaluate rate limit quota distributively in Redis for the specific client IP
	limiterCtx, err := r.limiterInstance.Get(context.Background(), ip)
	if err != nil {
		return false, err
	}

	// 2. Mengembalikan true jika permintaan belum melampaui limit (tidak Reached)
	// 2. Return true if request count has not hit the limit yet (not Reached)
	return !limiterCtx.Reached, nil
}
