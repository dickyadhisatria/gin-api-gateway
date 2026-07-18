package proxy

import (
	"context"
	"errors"
	"gin-api-gateway/internal/domain/service"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/sony/gobreaker"
)

type serviceProxy struct {
	reverseProxy *httputil.ReverseProxy
	breaker      *gobreaker.CircuitBreaker
}

type httpProxy struct {
	proxies sync.Map // key: host string, value: *serviceProxy
}

// NewHTTPProxy membuat instansi baru dari client proxy HTTP.
// NewHTTPProxy constructs a new instance of HTTP proxy client.
//
// Returns:
//   - service.ProxyClient: Implementasi antarmuka ProxyClient / ProxyClient interface implementation.
func NewHTTPProxy() service.ProxyClient {
	return &httpProxy{}
}

// getServiceProxy mengambil proxy dan circuit breaker dari cache atau membuatnya baru jika belum terdaftar.
// getServiceProxy retrieves the proxy and circuit breaker from cache, or constructs them if not registered.
//
// Parameters:
//   - remote (*url.URL): URL target mikroservis hilir / Downstream target microservice URL.
//
// Returns:
//   - *serviceProxy: Paket reverse proxy dan circuit breaker terikat / Bound reverse proxy and circuit breaker package.
func (p *httpProxy) getServiceProxy(remote *url.URL) *serviceProxy {
	host := remote.Host
	if val, ok := p.proxies.Load(host); ok {
		return val.(*serviceProxy)
	}

	// 1. Menginisialisasi ReverseProxy bawaan Go yang dioptimalkan untuk host tujuan
	// 1. Initialize the optimized Go standard ReverseProxy for target host
	rp := httputil.NewSingleHostReverseProxy(remote)
	
	// 1a. Mengoptimalkan Connection Pooling (Transport) untuk concurrency skala tinggi
	// 1a. Optimize Connection Pooling (Transport settings) for high-scale concurrency
	rp.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          1000,
		MaxIdleConnsPerHost:   100, // Meningkatkan limit default Go (2) untuk mencegah soket TIME_WAIT / Increase Go default (2) to prevent TIME_WAIT sockets
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	
	// 2. Menambahkan custom ErrorHandler untuk log detail error proxy dan respon JSON yang rapi
	// 2. Attach custom ErrorHandler for proxy failure logging and standard JSON error response
	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		slog.Error("Reverse proxy error", "error", err, "url", r.URL.String())
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, context.DeadlineExceeded) {
			w.WriteHeader(http.StatusGatewayTimeout)
			w.Write([]byte(`{"error":"Gateway Timeout: Downstream service took too long to respond"}`))
		} else {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(`{"error":"Bad Gateway: Failed to reach downstream service"}`))
		}
	}

	// 3. Mengatur konfigurasi Circuit Breaker untuk host tersebut
	// 3. Configure the Circuit Breaker parameters for the host
	settings := gobreaker.Settings{
		Name:        "Proxy-to-" + host,
		MaxRequests: 3,
		Interval:    5 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	}
	cb := gobreaker.NewCircuitBreaker(settings)

	sp := &serviceProxy{
		reverseProxy: rp,
		breaker:      cb,
	}

	actual, _ := p.proxies.LoadOrStore(host, sp)
	return actual.(*serviceProxy)
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader menangkap status kode respon sebelum dikirim ke penulis respon asli.
// WriteHeader intercepts response status code before writing to the original response writer.
//
// Parameters:
//   - statusCode (int): Kode status HTTP / HTTP status code.
func (w *statusResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write menulis data bodi respon dan memastikan status 200 diset jika belum diatur.
// Write writes response body data and defaults status to 200 OK if not explicitly set yet.
//
// Parameters:
//   - b ([]byte): Data bodi respon / Response body bytes.
//
// Returns:
//   - int: Jumlah byte tertulis / Written bytes count.
//   - error: Error penulisan / Writing error.
func (w *statusResponseWriter) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

// Serve melakukan perutean request ke target downstream di bawah proteksi Circuit Breaker.
// Serve routes the incoming request to the downstream target under Circuit Breaker context.
//
// Parameters:
//   - rw (http.ResponseWriter): Penulis respon HTTP / HTTP response writer.
//   - req (*http.Request): Objek request HTTP / HTTP request object.
//   - target (string): URL tujuan microservice / Downstream microservice target URL.
//
// Returns:
//   - error: Error dari pemanggilan proxy atau status sekring terbuka / Error from proxy dispatch or open circuit state.
func (p *httpProxy) Serve(rw http.ResponseWriter, req *http.Request, target string) error {
	// 1. Mengurai alamat target URL menjadi objek URL yang valid
	// 1. Parse target URL string into a valid URL object
	remote, err := url.Parse(target)
	if err != nil {
		return err
	}

	// 2. Mendapatkan cached serviceProxy (ReverseProxy & CircuitBreaker) untuk host target
	// 2. Fetch the cached serviceProxy package (ReverseProxy & CircuitBreaker) for the target host
	sp := p.getServiceProxy(remote)

	// 3. Membatasi durasi request ke microservice hilir dengan timeout 5 detik untuk mencegah goroutine leak
	// 3. Enforce a 5-second context timeout on microservice call to prevent gateway goroutine leakage
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	// 4. Menjalankan proses reverse proxy di bawah pengawasan Circuit Breaker (gobreaker)
	// 4. Run reverse proxy execution under the supervision of Circuit Breaker (gobreaker)
	_, errCB := sp.breaker.Execute(func() (interface{}, error) {
		// 4a. Membungkus ResponseWriter asli untuk menangkap status kode respon HTTP dari downstream
		// 4a. Intercept original ResponseWriter to capture HTTP response status from downstream
		writer := &statusResponseWriter{ResponseWriter: rw, statusCode: 0}

		// 4b. Melakukan pemanggilan proxy HTTP menggunakan instance cached
		// 4b. Forward HTTP request using the cached reverse proxy instance
		sp.reverseProxy.ServeHTTP(writer, req)

		// 4c. Mengembalikan error ke circuit breaker jika downstream mengembalikan kegagalan server (5xx)
		// 4c. Report error to Circuit Breaker if downstream service returns a server failure (5xx)
		if writer.statusCode >= 500 {
			return nil, errors.New("downstream service returned 5xx")
		}

		// 4d. Memeriksa apakah request terputus karena batas waktu (timeout)
		// 4d. Evaluate if microservice request timed out
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		return nil, nil
	})

	// 5. Jika eksekusi gagal atau sekring dalam status terbuka (Open), kembalikan error terkait
	// 5. If execution fails or circuit breaker state is Open/Half-Open, return the error
	if errCB != nil {
		return errCB
	}

	return nil
}
