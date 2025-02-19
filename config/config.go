package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	MemoryCacheSize int
	UseMemcached    bool
	MemcachedADDR   string
	RedisAddr       string
	RedisPassword   string
	DatabaseDSN     string
	LogLevel        string
}

func LoadConfig() *Config {
	// Load environment variables from .env if it exists
	_ = godotenv.Load()

	memoryCacheSize, err := strconv.Atoi(getEnv("MEMORY_CACHE_SIZE", "1000"))
	if err != nil {
		log.Fatalf("Error parsing MEMORY_CACHE_SIZE: %v", err)
	}

	return &Config{
		MemoryCacheSize: memoryCacheSize,
		UseMemcached:    getEnvBool("USE_MEMCACHED", false),
		MemcachedADDR:   getEnv("MEMCACHED_ADDR", "localhost:11211"),
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
