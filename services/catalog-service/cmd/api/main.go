package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

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

	if os.Getenv("PLATFORM_LAB_CPU_PRESSURE") == "true" {
		productListHandler = withLabCPUPressure(productListHandler)
	}

	if os.Getenv("PLATFORM_LAB_MEMORY_GROWTH") == "true" {
		productListHandler = withLabMemoryGrowth(productListHandler)
	}

	if os.Getenv("PLATFORM_LAB_OOM_PRESSURE") == "true" {
		productListHandler = withLabOOMPressure(productListHandler)
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

		f, err := os.OpenFile(
			labDiskGrowthFile,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			0o644,
		)
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

const labCPUPressureDuration = 500 * time.Millisecond

// withLabCPUPressure is a dormant training-only fault hook for Lesson 39.
// When armed, each controlled product-list request performs bounded CPU work
// before the normal Catalog handler runs. This creates a workload -> process ->
// CPU -> latency relationship, and CPU falls when the workload stops.
func withLabCPUPressure(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deadline := time.Now().Add(labCPUPressureDuration)
		sum := sha256.Sum256([]byte(r.URL.Path))

		for time.Now().Before(deadline) {
			sum = sha256.Sum256(sum[:])
		}

		log.Printf(
			"lab cpu-pressure: completed bounded expensive processing for %s checksum=%x",
			r.URL.Path,
			sum[:4],
		)

		next.ServeHTTP(w, r)
	}
}

const (
	labMemoryGrowthBytes       = 512 * 1024
	labMemoryGrowthMaximumSize = 128 * 1024 * 1024
)

var (
	labMemoryGrowthMu       sync.Mutex
	labMemoryGrowthRetained [][]byte
	labMemoryGrowthTotal    int
)

// withLabMemoryGrowth is a dormant training-only fault hook for Lesson 40.
// Each controlled product-list request retains a bounded block of memory.
// Because references to those blocks remain reachable, the Go garbage
// collector cannot reclaim them. Repeated workload therefore produces
// observable retained-memory growth while the Catalog service remains healthy.
// The hard cap keeps the local training scenario safe.
func withLabMemoryGrowth(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		labMemoryGrowthMu.Lock()
		defer labMemoryGrowthMu.Unlock()

		if labMemoryGrowthTotal >= labMemoryGrowthMaximumSize {
			return
		}

		remaining := labMemoryGrowthMaximumSize - labMemoryGrowthTotal
		allocationSize := labMemoryGrowthBytes

		if remaining < allocationSize {
			allocationSize = remaining
		}

		block := make([]byte, allocationSize)

		// Touch every page so the allocation becomes observable as real
		// resident memory instead of remaining only virtually allocated.
		for i := 0; i < len(block); i += 4096 {
			block[i] = byte((labMemoryGrowthTotal + i) % 251)
		}

		labMemoryGrowthRetained = append(
			labMemoryGrowthRetained,
			block,
		)

		labMemoryGrowthTotal += len(block)

		log.Printf(
			"lab memory-growth: retained %d KiB; total retained=%d MiB",
			len(block)/1024,
			labMemoryGrowthTotal/(1024*1024),
		)
	}
}

const labOOMPressureBlockSize = 4 * 1024 * 1024

var (
	labOOMPressureMu       sync.Mutex
	labOOMPressureRetained [][]byte
)

// withLabOOMPressure is a dormant training-only fault hook for Lesson 41.
// A controlled product-list request starts sustained retained-memory pressure.
// The Lab CLI places Catalog inside a strict container memory limit, so the
// process eventually crosses that real cgroup boundary and is OOM killed.
//
// This behavior is inactive during normal Platform Lab operation and is enabled
// only when PLATFORM_LAB_OOM_PRESSURE=true.
func withLabOOMPressure(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		labOOMPressureMu.Lock()
		defer labOOMPressureMu.Unlock()

		log.Printf("lab oom-pressure: controlled memory pressure started")

		for {
			block := make([]byte, labOOMPressureBlockSize)

			// Touch every page so the allocation becomes resident memory and
			// contributes to the container's real cgroup memory usage.
			for i := 0; i < len(block); i += 4096 {
				block[i] = byte((len(labOOMPressureRetained) + i) % 251)
			}

			labOOMPressureRetained = append(
				labOOMPressureRetained,
				block,
			)

			log.Printf(
				"lab oom-pressure: retained another %d MiB; total=%d MiB",
				labOOMPressureBlockSize/(1024*1024),
				(len(labOOMPressureRetained)*labOOMPressureBlockSize)/(1024*1024),
			)

			time.Sleep(100 * time.Millisecond)
		}
	}
}
