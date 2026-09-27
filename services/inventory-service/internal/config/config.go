package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr       string
	DatabaseURL    string
	KafkaBrokers   []string
	OrdersTopic    string
	InventoryTopic string
	ConsumerGroup  string
	OutboxPoll     time.Duration
}

func Load() Config {
	return Config{
		HTTPAddr:       env("HTTP_ADDR", ":8084"),
		DatabaseURL:    env("DATABASE_URL", "postgres://platform:platform@localhost:5432/inventory_db?sslmode=disable"),
		KafkaBrokers:   strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ","),
		OrdersTopic:    env("ORDERS_TOPIC", "orders"),
		InventoryTopic: env("INVENTORY_TOPIC", "inventory"),
		ConsumerGroup:  env("KAFKA_CONSUMER_GROUP", "inventory-service"),
		OutboxPoll:     time.Second,
	}
}

func env(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
