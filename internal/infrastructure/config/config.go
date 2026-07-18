package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config menampung data konfigurasi global aplikasi yang dibaca dari environment.
// Config holds the application global configuration read from variables.
type Config struct {
	Port           string // Port gateway (contoh: :3000) / Gateway listening port (e.g. :3000)
	JwtSecret      string // Kunci rahasia validasi JWT / JWT signing secret key
	UserService    string // URL downstream microservice user / Downstream user microservice URL
	AuthService    string // URL downstream microservice auth / Downstream auth microservice URL
	ProductService string // URL downstream microservice product / Downstream product microservice URL
	OrderService   string // URL downstream microservice order / Downstream order microservice URL
	RedisURL       string // Host dan port Redis (contoh: localhost:6379) / Redis host and port (e.g. localhost:6379)
	RedisPassword  string // Kata sandi otentikasi Redis / Redis authentication password
	RedisDB        int    // Indeks DB Redis / Redis database index number
}

// getEnv membaca variabel lingkungan atau mengembalikan nilai default jika kosong.
// getEnv reads environment variable or returns default fallback if empty.
//
// Parameters:
//   - key (string): Nama variabel lingkungan / Environment variable name.
//   - defaultValue (string): Nilai default cadangan / Fallback default value.
//
// Returns:
//   - string: Nilai variabel lingkungan / Env variable value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// LoadConfig memuat konfigurasi dari file .env lokal atau variabel lingkungan OS.
// LoadConfig initializes configuration properties from local .env or system environment.
//
// Returns:
//   - Config: Struktur konfigurasi terisi data / Loaded config data struct.
func LoadConfig() Config {
	// 1. Memuat variabel lingkungan dari file .env jika tersedia
	// 1. Load variables from local .env file if present
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found. Using default environment variables.")
	}

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	// 2. Mengisi field struktur Config dari variabel lingkungan dengan nilai default cadangan
	// 2. Map env variables into Config attributes with default values fallback
	return Config{
		Port:           getEnv("PORT", ":8080"),
		JwtSecret:      getEnv("JWT_SECRET", "default-jwt-secret"),
		AuthService:    getEnv("AUTH_SERVICE_URL", "http://localhost:3001"),
		UserService:    getEnv("USER_SERVICE_URL", "http://localhost:3002"),
		ProductService: getEnv("PRODUCT_SERVICE_URL", "http://localhost:3003"),
		OrderService:   getEnv("ORDER_SERVICE_URL", "http://localhost:3004"),
		RedisURL:       getEnv("REDIS_URL", ""),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		RedisDB:        redisDB,
	}
}
