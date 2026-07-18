package usecase

import (
	"gin-api-gateway/internal/domain/service"
	"net/http"
)

// ProxyUseCase mendefinisikan kasus penggunaan untuk merutekan permintaan HTTP.
// ProxyUseCase defines the use case for routing HTTP requests.
type ProxyUseCase interface {
	// Execute menjalankan perutean request ke target downstream.
	// Execute performs request routing to the downstream target.
	//
	// Parameters:
	//   - rw (http.ResponseWriter): Penulis respon HTTP / HTTP response writer.
	//   - req (*http.Request): Objek request HTTP / HTTP request object.
	//   - target (string): URL target tujuan / Downstream target URL.
	//
	// Returns:
	//   - error: Error eksekusi proxy / Proxy execution error.
	Execute(rw http.ResponseWriter, req *http.Request, target string) error
}

type proxyUseCase struct {
	proxyClient service.ProxyClient
}

// NewProxyUseCase membuat instansi baru dari ProxyUseCase.
// NewProxyUseCase constructs a new instance of ProxyUseCase.
//
// Parameters:
//   - client (service.ProxyClient): Instansi adapter proxy client / Proxy client adapter instance.
//
// Returns:
//   - ProxyUseCase: Kasus penggunaan proxy yang siap pakai / Active ProxyUseCase instance.
func NewProxyUseCase(client service.ProxyClient) ProxyUseCase {
	return &proxyUseCase{proxyClient: client}
}

// Execute menjalankan perutean request ke target downstream.
// Execute performs request routing to the downstream target.
//
// Parameters:
//   - rw (http.ResponseWriter): Penulis respon HTTP / HTTP response writer.
//   - req (*http.Request): Objek request HTTP / HTTP request object.
//   - target (string): URL target tujuan / Downstream target URL.
//
// Returns:
//   - error: Error eksekusi proxy / Proxy execution error.
func (uc *proxyUseCase) Execute(rw http.ResponseWriter, req *http.Request, target string) error {
	// 1. Meneruskan permintaan HTTP ke target downstream menggunakan adapter ProxyClient
	// 1. Forward the HTTP request to the downstream target using the ProxyClient adapter.
	return uc.proxyClient.Serve(rw, req, target)
}
