package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/gopher-opsx/platform-lab/services/inventory-service/internal/config"
	kafkainfra "github.com/gopher-opsx/platform-lab/services/inventory-service/internal/infrastructure/kafka"
	"github.com/gopher-opsx/platform-lab/services/inventory-service/internal/infrastructure/postgres"
	"github.com/gopher-opsx/platform-lab/services/inventory-service/internal/metrics"
	"github.com/gopher-opsx/platform-lab/services/inventory-service/internal/outbox"
	"github.com/gopher-opsx/platform-lab/services/inventory-service/internal/service"
	"github.com/gopher-opsx/platform-lab/services/inventory-service/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	cfg := config.Load()
	shutdownTelemetry, err := telemetry.Start(context.Background(), "inventory-service")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = shutdownTelemetry(context.Background()) }()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	db, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	svc := service.New(postgres.NewInventoryRepository(db), cfg.InventoryTopic)
	pub := kafkainfra.NewPublisher(cfg.KafkaBrokers)
	defer pub.Close()
	go outbox.New(postgres.NewOutboxRepository(db), pub, cfg.OutboxPoll).Run(ctx)
	consumer := kafkainfra.NewConsumer(cfg.KafkaBrokers, cfg.OrdersTopic, cfg.ConsumerGroup, svc)
	defer consumer.Close()
	go consumer.Run(ctx)
	mux := http.NewServeMux()
	metricCollector := metrics.New("inventory-service")
	mux.Handle("/metrics", metricCollector.Handler())
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(r.Context()); err != nil {
			http.Error(w, "database not ready", 503)
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ready"))
	})
	s := &http.Server{Addr: cfg.HTTPAddr, Handler: otelhttp.NewHandler(metricCollector.Middleware(mux), "inventory-service.http")}
	go func() {
		log.Printf("inventory-service listening on %s", cfg.HTTPAddr)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	_ = s.Shutdown(context.Background())
}
