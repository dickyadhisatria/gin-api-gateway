# Gin API Gateway

API Gateway yang tangguh, aman, dan berkinerja tinggi yang dibangun dari nol menggunakan **Go** dan framework **Gin Gonic**. Proyek ini dirancang dengan presisi arsitektural menggunakan prinsip **Clean Architecture**, yang menegakkan pemisahan kepentingan secara ketat (separation of concerns), aturan dependensi, dan inversi dependensi.

API Gateway ini berfungsi sebagai pintu masuk tunggal untuk arsitektur microservices, yang mengelola perutean lalu lintas (traffic routing), keamanan, dan ketahanan sistem (resilience).

## 🚀 Fitur

- **Implementasi Clean Architecture**: Pembagian ketat ke dalam lapisan Domain, Use Case, Adapter, dan Infrastruktur. Sepenuhnya terlepas dari framework tertentu untuk mempermudah pengujian (testability).
- **Routing Reverse Proxy**: Merutekan permintaan masuk secara dinamis ke microservice di hilir.
- **Connection Pooling Teroptimasi**: Konfigurasi `http.Transport` khusus (`MaxIdleConnsPerHost: 100`) untuk mencegah penumpukan soket `TIME_WAIT` dan latensi pembukaan koneksi baru di bawah lalu lintas tinggi.
- **Pembatasan Laju Terdistribusi (Distributed Rate Limiting)**: Mendukung penyimpanan lokal memori (in-memory) dan **Redis** menggunakan `ulule/limiter` untuk kemudahan scaling horizontal. Otomatis kembali ke memori jika konfigurasi Redis kosong.
- **Keamanan & Autentikasi**: Mengimplementasikan autentikasi dan otorisasi berbasis JWT untuk akses aman ke microservices.
- **Toleransi Kesalahan (Circuit Breaker)**: Mencegah kegagalan beruntun menggunakan `sony/gobreaker` yang dipetakan secara dinamis dan di-cache per host tujuan.
- **Proteksi Batas Waktu Hilir (Downstream Timeout Guard)**: Menerapkan batas waktu (timeout) context sebesar 5 detik pada layanan mikro hilir guna menghindari kebocoran (leak) goroutine di gateway.
- **Manajemen Konfigurasi**: Tanpa hardcoding nilai, sepenuhnya dikendalikan oleh variabel lingkungan (`.env`).
- **Pencatatan Log & Pemantauan**: Pencatatan log permintaan berformat JSON terstruktur (`slog`) dengan Request ID aman secara kriptografis dan standarisasi respon error gateway dalam format JSON.

## 📁 Struktur Proyek

```text
gin-api-gateway/
├── cmd/
│   └── api/
│       └── main.go                 # Composition Root (Dependency Injection)
├── internal/
│   ├── domain/
│   │   ├── model/                  # Entitas domain (misalnya klaim pengguna)
│   │   └── service/                # Definisi antarmuka/interface domain
│   ├── usecase/                    # Kasus penggunaan bisnis (misalnya ProxyUseCase)
│   ├── adapter/
│   │   ├── auth/                   # Adapter verifikasi JWT
│   │   ├── ratelimit/              # Adapter rate limiter memori & Redis
│   │   ├── proxy/                  # Adapter reverse proxy httputil + gobreaker
│   │   └── delivery/
│   │       └── http/
│   │           ├── handler/        # HTTP controller/handler
│   │           └── middleware/     # Middleware HTTP (logging, rate limiting, dan auth)
│   └── infrastructure/
│       └── config/                 # Pemuat variabel lingkungan
├── .env                            # Konfigurasi lokal rahasia (diabaikan Git)
├── .gitignore                      # Mengatur file yang diabaikan oleh Git
├── go.mod                          # Konfigurasi modul Go
├── go.sum                          # Checksum modul Go
├── k6.js                           # Skrip pengujian beban K6 (lokal, gitignored)
└── README.md                       # Dokumentasi
```

## 🛠️ Memulai

### 1. Prasyarat

Pastikan Anda telah menginstal **Go 1.26+** di komputer Anda.

### 2. Instalasi

Klon repositori dan unduh dependensi yang diperlukan:

```bash
git clone https://github.com/dickyadhisatria/gin-api-gateway
cd gin-api-gateway
go mod tidy
```

### 3. Pengaturan Variabel Lingkungan

Buat file `.env` di direktori utama (root):

```text
PORT=:3000
JWT_SECRET=your-secure-jwt-secret
# Downstream microservices URLs (ganti dengan URL layanan Anda yang sebenarnya)
AUTH_SERVICE_URL=http://localhost:3001
USER_SERVICE_URL=http://localhost:3002
PRODUCT_SERVICE_URL=http://localhost:3003
ORDER_SERVICE_URL=http://localhost:3004

# Konfigurasi Redis (Opsional, otomatis kembali ke memori jika kosong)
REDIS_URL=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

### 4. Menjalankan Aplikasi

Mulai API Gateway:

```bash
go run cmd/api/main.go
```

Gateway akan mulai berjalan di port `:3000`.

### 5. Menjalankan Pengujian (Tests)

Jalankan pengujian unit:

```bash
go test -v ./...
```

## 🧪 Menguji Gateway

Uji autentikasi JWT menggunakan `cURL`:

```bash
# Perintah ini akan mengembalikan status 401 Unauthorized
curl -i http://localhost:3000/api/v1/user/profile

# Perintah ini akan berhasil diarahkan ke user service Anda
curl -i -H "Authorization: Bearer <token-jwt-anda>" http://localhost:3000/api/v1/user/profile
```
