package service

import (
	"gin-api-gateway/internal/domain/model"
	"net/http"
)

// Authenticator mendefinisikan kontrak untuk memvalidasi token keamanan.
// Authenticator defines the contract for validating security tokens.
type Authenticator interface {
	// Verify memverifikasi token JWT dan mengembalikan informasi klaim pengguna.
	// Verify validates the JWT token and returns user claims details.
	//
	// Parameters:
	//   - token (string): Token JWT mentah / Raw JWT token string.
	//
	// Returns:
	//   - *model.UserClaims: Data klaim pengguna / User claims entity.
	//   - error: Error verifikasi jika token tidak valid / Verification error if token is invalid.
	Verify(token string) (*model.UserClaims, error)
}

// RateLimiter mendefinisikan kontrak untuk pembatasan laju permintaan HTTP.
// RateLimiter defines the contract for HTTP request rate limiting.
type RateLimiter interface {
	// Allow memeriksa apakah alamat IP klien masih berada dalam kuota yang diizinkan.
	// Allow checks if the client IP address is still within the permitted request quota.
	//
	// Parameters:
	//   - ip (string): Alamat IP klien / Client IP address.
	//
	// Returns:
	//   - bool: True jika diperbolehkan, False jika dibatasi / True if allowed, False if limited.
	//   - error: Error internal server / Internal server error.
	Allow(ip string) (bool, error)
}

// ProxyClient mendefinisikan kontrak untuk memforward request HTTP ke microservice hilir.
// ProxyClient defines the contract for forwarding HTTP requests to downstream microservices.
type ProxyClient interface {
	// Serve mengirimkan request HTTP ke URL tujuan dengan pengaman sekring (Circuit Breaker).
	// Serve dispatches the HTTP request to the target URL with Circuit Breaker protection.
	//
	// Parameters:
	//   - rw (http.ResponseWriter): Penulis respon HTTP / HTTP response writer.
	//   - req (*http.Request): Objek request HTTP masuk / Incoming HTTP request object.
	//   - target (string): URL microservice tujuan / Downstream microservice target URL.
	//
	// Returns:
	//   - error: Error dari pemanggilan proxy atau sekring terbuka / Error from proxy dispatch or tripped breaker.
	Serve(rw http.ResponseWriter, req *http.Request, target string) error
}
