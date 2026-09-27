package config

import "os"

type Config struct {
	HTTPAddr, CatalogURL, CartURL, OrderURL, AllowedOrigin string
}

func Load() Config {
	return Config{
		HTTPAddr: env("HTTP_ADDR", ":8080"), CatalogURL: env("CATALOG_SERVICE_URL", "http://localhost:8081"),
		CartURL: env("CART_SERVICE_URL", "http://localhost:8082"), OrderURL: env("ORDER_SERVICE_URL", "http://localhost:8083"),
		AllowedOrigin: env("ALLOWED_ORIGIN", "http://localhost:4200"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
