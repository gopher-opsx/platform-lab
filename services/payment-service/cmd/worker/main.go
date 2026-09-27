package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gopher-opsx/platform-lab/services/payment-service/internal/config"
	kafkainfra "github.com/gopher-opsx/platform-lab/services/payment-service/internal/infrastructure/kafka"
	"github.com/gopher-opsx/platform-lab/services/payment-service/internal/infrastructure/postgres"
	"github.com/gopher-opsx/platform-lab/services/payment-service/internal/metrics"
	"github.com/gopher-opsx/platform-lab/services/payment-service/internal/outbox"
	"github.com/gopher-opsx/platform-lab/services/payment-service/internal/service"
	"github.com/gopher-opsx/platform-lab/services/payment-service/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	cfg := config.Load()
	shutdownTelemetry, err := telemetry.Start(context.Background(), "payment-service")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = shutdownTelemetry(context.Background()) }()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbPool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	paymentRepository := postgres.NewPaymentRepository(dbPool)
	paymentService := service.NewPaymentService(paymentRepository, cfg.PaymentsTopic, cfg.MaxAuthCents)

	kafkaPublisher := kafkainfra.NewPublisher(cfg.KafkaBrokers)
	defer kafkaPublisher.Close()

	outboxRepository := postgres.NewOutboxRepository(dbPool)
	outboxPublisher := outbox.NewPublisher(outboxRepository, kafkaPublisher, cfg.OutboxPoll, cfg.OutboxBatchSize)

	kafkaConsumer := kafkainfra.NewConsumer(cfg.KafkaBrokers, cfg.InventoryTopic, cfg.ConsumerGroup, paymentService)
	defer kafkaConsumer.Close()

	go outboxPublisher.Run(ctx)
	go kafkaConsumer.Run(ctx)

	mux := http.NewServeMux()
	metricCollector := metrics.New("payment-service")
	mux.Handle("/metrics", metricCollector.Handler())
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := dbPool.Ping(r.Context()); err != nil {
			http.Error(w, "database not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	server := &http.Server{Addr: cfg.HTTPAddr, Handler: otelhttp.NewHandler(metricCollector.Middleware(mux), "payment-service.http"), ReadHeaderTimeout: 5 * time.Second}
	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("payment-service listening on %s", cfg.HTTPAddr)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("payment-service shutdown failed: %v", err)
		}
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(fmt.Errorf("payment-service failed: %w", err))
		}
	}
}
