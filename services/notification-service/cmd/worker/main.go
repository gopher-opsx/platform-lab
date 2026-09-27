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

	"github.com/gopher-opsx/platform-lab/services/notification-service/internal/config"
	kafkainfra "github.com/gopher-opsx/platform-lab/services/notification-service/internal/infrastructure/kafka"
	"github.com/gopher-opsx/platform-lab/services/notification-service/internal/infrastructure/postgres"
	"github.com/gopher-opsx/platform-lab/services/notification-service/internal/metrics"
	"github.com/gopher-opsx/platform-lab/services/notification-service/internal/service"
	"github.com/gopher-opsx/platform-lab/services/notification-service/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	cfg := config.Load()
	shutdownTelemetry, err := telemetry.Start(context.Background(), "notification-service")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = shutdownTelemetry(context.Background()) }()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	notificationService := service.NewNotificationService(postgres.NewNotificationRepository(pool))
	consumer := kafkainfra.NewConsumer(cfg.KafkaBrokers, cfg.OrdersTopic, cfg.ConsumerGroup, notificationService)
	defer consumer.Close()
	go consumer.Run(ctx)

	mux := http.NewServeMux()
	metricCollector := metrics.New("notification-service")
	mux.Handle("/metrics", metricCollector.Handler())
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			http.Error(w, "database not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: otelhttp.NewHandler(metricCollector.Middleware(mux), "notification-service.http"), ReadHeaderTimeout: 5 * time.Second}
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("notification-service listening on %s", cfg.HTTPAddr)
		serverErrors <- server.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("notification-service shutdown failed: %v", err)
		}
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(fmt.Errorf("notification-service failed: %w", err))
		}
	}
}
