package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	ApiKey         string
	UserService    string
	ProductService string
	OrderService   string
}

// LoadConfig membaca pengaturan dari file .env
func LoadConfig() Config {
	// Ambil file .env, jika gagal aplikasi akan memberikan peringatan
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, menggunakan pengaturan bawaan sistem")
	}

	return Config{
		Port:           getEnv("PORT", ":8080"),
		ApiKey:         getEnv("API_KEY", "kunci-rahasia-gateway-123"),
		UserService:    getEnv("USER_SERVICE_URL", "http://localhost:8081"),
		ProductService: getEnv("PRODUCT_SERVICE_URL", "http://localhost:8082"),
		OrderService:   getEnv("ORDER_SERVICE_URL", "http://localhost:8083"),
	}
}

// Fungsi pembantu untuk mengambil data env atau menggunakan nilai cadangan jika kosong
func getEnv(kunci, defaultValue string) string {
	if nilai := os.Getenv(kunci); nilai != "" {
		return nilai
	}
	return defaultValue
}
