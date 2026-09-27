package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gopher-opsx/platform-lab/services/catalog-service/internal/config"
	"github.com/gopher-opsx/platform-lab/services/catalog-service/internal/infrastructure/postgres"
	"github.com/gopher-opsx/platform-lab/services/catalog-service/internal/metrics"
	"github.com/gopher-opsx/platform-lab/services/catalog-service/internal/service"
	"github.com/gopher-opsx/platform-lab/services/catalog-service/internal/telemetry"
	httptransport "github.com/gopher-opsx/platform-lab/services/catalog-service/internal/transport/http"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	cfg := config.Load()
	shutdownTelemetry, err := telemetry.Start(context.Background(), "catalog-service")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = shutdownTelemetry(context.Background()) }()

	ctx := context.Background()

	dbPool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	productRepository := postgres.NewProductRepository(dbPool)
	catalogService := service.NewCatalogService(productRepository)
	catalogHandler := httptransport.NewCatalogHandler(catalogService)

	mux := http.NewServeMux()
	metricCollector := metrics.New("catalog-service")
	mux.Handle("/metrics", metricCollector.Handler())

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := dbPool.Ping(r.Context()); err != nil {
			http.Error(w, "database not ready", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	mux.HandleFunc("GET /products", catalogHandler.ListProducts)
	mux.HandleFunc("GET /products/{id}", catalogHandler.GetProduct)

	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: otelhttp.NewHandler(metricCollector.Middleware(mux), "catalog-service.http"),
	}

	log.Printf("catalog-service listening on %s", cfg.HTTPAddr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(fmt.Errorf("catalog-service failed: %w", err))
	}
}
