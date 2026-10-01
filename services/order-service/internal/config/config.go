package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPAddr               = ":8083"
	defaultDatabaseURL            = "postgres://platform:platform@localhost:5432/orders_db?sslmode=disable"
	defaultKafkaBrokers           = "localhost:9092"
	defaultOrdersTopic            = "orders"
	defaultPaymentsTopic          = "payments"
	defaultInventoryTopic         = "inventory"
	defaultPaymentsConsumerGroup  = "order-service-payments"
	defaultInventoryConsumerGroup = "order-service-inventory"
	defaultOutboxPoll             = time.Second
	defaultOutboxBatchSize        = 50
)

type Config struct {
	HTTPAddr               string
	DatabaseURL            string
	KafkaBrokers           []string
	OrdersTopic            string
	PaymentsTopic          string
	InventoryTopic         string
	PaymentsConsumerGroup  string
	InventoryConsumerGroup string
	OutboxPoll             time.Duration
	OutboxBatchSize        int
}

func Load() Config {
	return Config{
		HTTPAddr:               getEnv("HTTP_ADDR", defaultHTTPAddr),
		DatabaseURL:            getEnv("DATABASE_URL", defaultDatabaseURL),
		KafkaBrokers:           splitCSV(getEnv("KAFKA_BROKERS", defaultKafkaBrokers)),
		OrdersTopic:            getEnv("ORDERS_TOPIC", defaultOrdersTopic),
		PaymentsTopic:          getEnv("PAYMENTS_TOPIC", defaultPaymentsTopic),
		InventoryTopic:         getEnv("INVENTORY_TOPIC", defaultInventoryTopic),
		PaymentsConsumerGroup:  getEnv("PAYMENTS_CONSUMER_GROUP", defaultPaymentsConsumerGroup),
		InventoryConsumerGroup: getEnv("INVENTORY_CONSUMER_GROUP", defaultInventoryConsumerGroup),
		OutboxPoll:             getEnvDuration("OUTBOX_POLL_INTERVAL", defaultOutboxPoll),
		OutboxBatchSize:        getEnvInt("OUTBOX_BATCH_SIZE", defaultOutboxBatchSize),
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
