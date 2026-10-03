package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

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

	productListHandler := http.HandlerFunc(catalogHandler.ListProducts)
	if os.Getenv("PLATFORM_LAB_DISK_GROWTH") == "true" {
		productListHandler = withLabDiskGrowth(productListHandler)
	}

	mux.Handle("GET /products", productListHandler)
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

const (
	labDiskGrowthFile        = "/tmp/platform-lab-disk-growth.bin"
	labDiskGrowthBytes       = 256 * 1024
	labDiskGrowthMaximumSize = 128 * 1024 * 1024
)

var labDiskGrowthMu sync.Mutex

// withLabDiskGrowth is a dormant training-only fault hook. It is enabled only
// by the Lesson 38 Compose override. Each successful controlled product-list
// request appends a deterministic amount of data to the container writable
// layer, giving students a real workload -> file -> container-storage causal
// chain to investigate. The cap keeps the local lab safe.
func withLabDiskGrowth(next http.Handler) http.HandlerFunc {
	payload := make([]byte, labDiskGrowthBytes)

	return func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		labDiskGrowthMu.Lock()
		defer labDiskGrowthMu.Unlock()

		info, err := os.Stat(labDiskGrowthFile)
		if err != nil && !os.IsNotExist(err) {
			log.Printf("lab disk-growth: inspect growth file: %v", err)
			return
		}
		if err == nil && info.Size() >= labDiskGrowthMaximumSize {
			return
		}

		f, err := os.OpenFile(labDiskGrowthFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			log.Printf("lab disk-growth: open growth file: %v", err)
			return
		}
		defer f.Close()

		if _, err := f.Write(payload); err != nil {
			log.Printf("lab disk-growth: append growth file: %v", err)
		}
	}
}
