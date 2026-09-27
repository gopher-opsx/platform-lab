package config

import (
	"os"
	"strconv"
	"time"
)

const (
	defaultHTTPAddr  = ":8082"
	defaultRedisAddr = "localhost:6379"
	defaultRedisDB   = 0
	defaultCartTTL   = 24 * time.Hour
)

type Config struct {
	HTTPAddr        string
	RedisAddr       string
	RedisPassword   string
	RedisTLSEnabled bool
	RedisDB         int
	CartTTL         time.Duration
}

func Load() Config {
	return Config{
		HTTPAddr:        getEnv("HTTP_ADDR", defaultHTTPAddr),
		RedisAddr:       getEnv("REDIS_ADDR", defaultRedisAddr),
		RedisPassword:   os.Getenv("REDIS_PASSWORD"),
		RedisTLSEnabled: getEnvBool("REDIS_TLS_ENABLED", false),
		RedisDB:         getEnvInt("REDIS_DB", defaultRedisDB),
		CartTTL:         getEnvDuration("CART_TTL", defaultCartTTL),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
