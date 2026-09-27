package config

import "os"

const (
	defaultHTTPAddr    = ":8081"
	defaultDatabaseURL = "postgres://platform:platform@localhost:5432/catalog_db?sslmode=disable"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
}

func Load() Config {
	return Config{
		HTTPAddr:    getEnv("HTTP_ADDR", defaultHTTPAddr),
		DatabaseURL: getEnv("DATABASE_URL", defaultDatabaseURL),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
