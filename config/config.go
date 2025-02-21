package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	MemoryCacheSize int
	RedisAddr       string
	RedisPassword   string
	DatabaseDSN     string
	LogLevel        string
}

func LoadConfig() *Config {
	// Load environment variables from .env if it exists
	_ = godotenv.Load()

	return &Config{
		MemoryCacheSize: getEnvInt("MEMORY_CACHE_SIZE", 100),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		DatabaseDSN:     getEnv("DATABASE_DSN", "postgres://user:password@localhost:5432/mydb"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
	}
}

// Helper function to get value from ENV with default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Helper function to get int value from ENV with default value
func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			log.Fatalf("Error parsing %s: %v", key, err)
		}
		return parsed
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			log.Fatalf("Error parsing %s: %v", key, err)
		}
		return parsed
	}
	return defaultValue
}
