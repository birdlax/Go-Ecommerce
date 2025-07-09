// config/config.go
package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config struct จะเก็บค่า Config ทั้งหมดของแอปพลิเคชัน
type Config struct {
	AzureConnectionString string
	PostgresDSN           string
	ImageBaseURL          string
}

// LoadConfig จะทำหน้าที่โหลด .env และคืนค่า Config struct ที่พร้อมใช้งาน
func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// ดึงค่าจาก Env Var
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	return &Config{
		AzureConnectionString: os.Getenv("AZURE_STORAGE_CONNECTION_STRING"),
		PostgresDSN:           dsn,
		ImageBaseURL:          os.Getenv("AZURE_STORAGE_BASE_URL"),
	}
}
