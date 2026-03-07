package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string
	DBUser  string
	DBPass  string
	DBHost  string
	DBPort  string
	DBName  string
}

func LoadConfig() *Config {
	err := godotenv.Load("config/.env")
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	return &Config{
		AppPort: getEnv("APP_PORT", "8080"),
		DBUser:  getEnv("DB_USERNAME", "root"),
		DBPass:  getEnv("DB_PASSWORD", ""),
		DBHost:  getEnv("DB_HOST", "127.0.0.1"),
		DBPort:  getEnv("DB_PORT", "3306"),
		DBName:  getEnv("DB_DATABASE", "spbu_go"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
