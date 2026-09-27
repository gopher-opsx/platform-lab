package config

import (
	"os"
	"strings"
)

type Config struct {
	HTTPAddr      string
	DatabaseURL   string
	KafkaBrokers  []string
	OrdersTopic   string
	ConsumerGroup string
}

func Load() Config {
	return Config{
		HTTPAddr:      getEnv("HTTP_ADDR", ":8086"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://platform:platform@localhost:5432/notifications_db?sslmode=disable"),
		KafkaBrokers:  splitCSV(getEnv("KAFKA_BROKERS", "localhost:9092")),
		OrdersTopic:   getEnv("ORDERS_TOPIC", "orders"),
		ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "notification-service-v1"),
	}
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			result = append(result, value)
		}
	}
	return result
}
