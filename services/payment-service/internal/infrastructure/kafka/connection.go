package kafka

import (
	"crypto/tls"
	"os"
	"strings"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
)

const (
	securityProtocolPlaintext = "PLAINTEXT"
	securityProtocolSASLSSL   = "SASL_SSL"
	saslMechanismPlain        = "PLAIN"
)

func newKafkaDialer() *kafkago.Dialer {
	tlsConfig, mechanism := kafkaSecurity()
	return &kafkago.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		TLS:           tlsConfig,
		SASLMechanism: mechanism,
	}
}

func newKafkaTransport() *kafkago.Transport {
	tlsConfig, mechanism := kafkaSecurity()
	return &kafkago.Transport{
		TLS:  tlsConfig,
		SASL: mechanism,
	}
}

func kafkaSecurity() (*tls.Config, sasl.Mechanism) {
	protocol := strings.ToUpper(strings.TrimSpace(envOrDefault("KAFKA_SECURITY_PROTOCOL", securityProtocolPlaintext)))
	mechanismName := strings.ToUpper(strings.TrimSpace(envOrDefault("KAFKA_SASL_MECHANISM", saslMechanismPlain)))

	if protocol == securityProtocolPlaintext {
		return nil, nil
	}

	if protocol != securityProtocolSASLSSL {
		panic("unsupported KAFKA_SECURITY_PROTOCOL: " + protocol)
	}
	if mechanismName != saslMechanismPlain {
		panic("unsupported KAFKA_SASL_MECHANISM: " + mechanismName)
	}

	return &tls.Config{MinVersion: tls.VersionTLS12}, plain.Mechanism{
		Username: os.Getenv("KAFKA_SASL_USERNAME"),
		Password: os.Getenv("KAFKA_SASL_PASSWORD"),
	}
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
