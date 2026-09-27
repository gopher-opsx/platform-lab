package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/gopher-opsx/platform-lab/services/cart-service/internal/config"
	redisrepo "github.com/gopher-opsx/platform-lab/services/cart-service/internal/infrastructure/redis"
	"github.com/gopher-opsx/platform-lab/services/cart-service/internal/metrics"
	"github.com/gopher-opsx/platform-lab/services/cart-service/internal/service"
	"github.com/gopher-opsx/platform-lab/services/cart-service/internal/telemetry"
	httptransport "github.com/gopher-opsx/platform-lab/services/cart-service/internal/transport/http"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	cfg := config.Load()
	shutdownTelemetry, err := telemetry.Start(context.Background(), "cart-service")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = shutdownTelemetry(context.Background()) }()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	redisOptions := &redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}
	if cfg.RedisTLSEnabled {
		redisOptions.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	redisClient := redis.NewClient(redisOptions)
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal(fmt.Errorf("connect to redis: %w", err))
	}

	cartRepository := redisrepo.NewCartRepository(redisClient, cfg.CartTTL)
	cartService := service.NewCartService(cartRepository)
	cartHandler := httptransport.NewCartHandler(cartService)

	mux := http.NewServeMux()
	metricCollector := metrics.New("cart-service")
	mux.Handle("/metrics", metricCollector.Handler())

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := redisClient.Ping(r.Context()).Err(); err != nil {
			http.Error(w, "redis not ready", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	mux.HandleFunc("GET /cart", cartHandler.GetCart)
	mux.HandleFunc("POST /cart/items", cartHandler.AddItem)
	mux.HandleFunc("PATCH /cart/items/{productId}", cartHandler.UpdateItem)
	mux.HandleFunc("DELETE /cart/items/{productId}", cartHandler.RemoveItem)
	mux.HandleFunc("DELETE /cart", cartHandler.ClearCart)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           otelhttp.NewHandler(metricCollector.Middleware(mux), "cart-service.http"),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("cart-service listening on %s", cfg.HTTPAddr)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("cart-service shutdown failed: %v", err)
		}
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(fmt.Errorf("cart-service failed: %w", err))
		}
	}
}
