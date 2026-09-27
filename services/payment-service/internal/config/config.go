package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPAddr        = ":8085"
	defaultDatabaseURL     = "postgres://platform:platform@localhost:5432/payments_db?sslmode=disable"
	defaultKafkaBrokers    = "localhost:9092"
	defaultInventoryTopic  = "inventory"
	defaultPaymentsTopic   = "payments"
	defaultConsumerGroup   = "payment-service"
	defaultMaxAuthCents    = int64(500000)
	defaultOutboxPoll      = time.Second
	defaultOutboxBatchSize = 50
)

type Config struct {
	HTTPAddr        string
	DatabaseURL     string
	KafkaBrokers    []string
	InventoryTopic  string
	PaymentsTopic   string
	ConsumerGroup   string
	MaxAuthCents    int64
	OutboxPoll      time.Duration
	OutboxBatchSize int
}

func Load() Config {
	return Config{
		HTTPAddr:        getEnv("HTTP_ADDR", defaultHTTPAddr),
		DatabaseURL:     getEnv("DATABASE_URL", defaultDatabaseURL),
		KafkaBrokers:    splitCSV(getEnv("KAFKA_BROKERS", defaultKafkaBrokers)),
		InventoryTopic:  getEnv("INVENTORY_TOPIC", defaultInventoryTopic),
		PaymentsTopic:   getEnv("PAYMENTS_TOPIC", defaultPaymentsTopic),
		ConsumerGroup:   getEnv("KAFKA_CONSUMER_GROUP", defaultConsumerGroup),
		MaxAuthCents:    getEnvInt64("PAYMENT_MAX_AUTH_CENTS", defaultMaxAuthCents),
		OutboxPoll:      getEnvDuration("OUTBOX_POLL_INTERVAL", defaultOutboxPoll),
		OutboxBatchSize: getEnvInt("OUTBOX_BATCH_SIZE", defaultOutboxBatchSize),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
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

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func getEnvInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
