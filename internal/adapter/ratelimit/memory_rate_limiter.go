package ratelimit

import (
	"context"
	"gin-api-gateway/internal/domain/service"
	"time"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

type memoryRateLimiter struct {
	limiterInstance *limiter.Limiter
}

// NewMemoryRateLimiter menginisialisasi instansi pembatas laju berbasis memori lokal.
// NewMemoryRateLimiter initializes an in-memory rate limiter instance.
//
// Parameters:
//   - limit (int64): Jumlah maksimum permintaan yang diizinkan / Maximum request count allowed.
//   - period (time.Duration): Periode waktu perputaran limit (misal: 1 menit) / Limit reset time period (e.g. 1 minute).
//
// Returns:
//   - service.RateLimiter: Implementasi antarmuka RateLimiter / RateLimiter interface implementation.
func NewMemoryRateLimiter(limit int64, period time.Duration) service.RateLimiter {
	rate := limiter.Rate{
		Period: period,
		Limit:  limit,
	}
	store := memory.NewStore()
	instance := limiter.New(store, rate)

	return &memoryRateLimiter{
		limiterInstance: instance,
	}
}

// Allow mengecek apakah alamat IP klien diperbolehkan mengirimkan request berdasarkan kuota.
// Allow checks if the client IP address is allowed to send request based on the quota.
//
// Parameters:
//   - ip (string): Alamat IP dari klien / Client IP address.
//
// Returns:
//   - bool: True jika diizinkan, False jika kuota habis / True if allowed, False if quota exceeded.
//   - error: Error internal dari tracker rate-limit / Internal error from the rate-limit tracker.
func (r *memoryRateLimiter) Allow(ip string) (bool, error) {
	// 1. Mengambil data limiter context untuk IP client tertentu secara sinkron menggunakan driver penyimpanan memori
	// 1. Retrieve the rate limit context for the specified client IP synchronously using the memory store driver
	limiterCtx, err := r.limiterInstance.Get(context.Background(), ip)
	if err != nil {
		return false, err
	}

	// 2. Memeriksa apakah jumlah permintaan dari IP tersebut telah melampaui batas kuota (Reached)
	// 2. Check if the request count from this IP has exceeded the quota limit (Reached)
	return !limiterCtx.Reached, nil
}
