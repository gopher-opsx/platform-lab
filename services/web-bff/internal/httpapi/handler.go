package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Handler struct {
	catalog, cart, order *url.URL
	client               *http.Client
	allowedOrigin        string
}

func New(catalogURL, cartURL, orderURL, allowedOrigin string) (*Handler, error) {
	catalog, err := url.Parse(catalogURL)
	if err != nil {
		return nil, err
	}
	cart, err := url.Parse(cartURL)
	if err != nil {
		return nil, err
	}
	order, err := url.Parse(orderURL)
	if err != nil {
		return nil, err
	}
	return &Handler{catalog: catalog, cart: cart, order: order, allowedOrigin: allowedOrigin, client: &http.Client{Timeout: 10 * time.Second, Transport: otelhttp.NewTransport(http.DefaultTransport)}}, nil
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", h.ready)
	mux.HandleFunc("GET /api/products", h.proxy(h.catalog, "/products"))
	mux.HandleFunc("GET /api/products/{id}", h.proxy(h.catalog, "/products/{id}"))
	mux.HandleFunc("GET /api/cart", h.proxy(h.cart, "/cart"))
	mux.HandleFunc("POST /api/cart/items", h.proxy(h.cart, "/cart/items"))
	mux.HandleFunc("PATCH /api/cart/items/{productId}", h.proxy(h.cart, "/cart/items/{productId}"))
	mux.HandleFunc("DELETE /api/cart/items/{productId}", h.proxy(h.cart, "/cart/items/{productId}"))
	mux.HandleFunc("DELETE /api/cart", h.proxy(h.cart, "/cart"))
	mux.HandleFunc("POST /api/orders", h.proxy(h.order, "/orders"))
	mux.HandleFunc("GET /api/orders", h.proxy(h.order, "/orders"))
	mux.HandleFunc("GET /api/orders/{id}", h.proxy(h.order, "/orders/{id}"))
	return h.cors(mux)
}

func (h *Handler) proxy(base *url.URL, pattern string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := pattern
		for _, key := range []string{"id", "productId"} {
			path = strings.ReplaceAll(path, "{"+key+"}", url.PathEscape(r.PathValue(key)))
		}
		target := *base
		target.Path = strings.TrimRight(base.Path, "/") + path
		target.RawQuery = r.URL.RawQuery
		request, err := http.NewRequestWithContext(r.Context(), r.Method, target.String(), r.Body)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "request creation failed"})
			return
		}
		copyHeader(request.Header, r.Header, "Content-Type")
		copyHeader(request.Header, r.Header, "Accept")
		copyHeader(request.Header, r.Header, "X-Customer-ID")
		copyHeader(request.Header, r.Header, "traceparent")
		response, err := h.client.Do(request)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "downstream service unavailable"})
			return
		}
		defer response.Body.Close()
		for _, key := range []string{"Content-Type", "Location"} {
			if value := response.Header.Get(key); value != "" {
				w.Header().Set(key, value)
			}
		}
		w.WriteHeader(response.StatusCode)
		_, _ = io.Copy(w, response.Body)
	}
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	checks := map[string]*url.URL{"catalog": h.catalog, "cart": h.cart, "order": h.order}
	failures := []string{}
	for name, base := range checks {
		target := *base
		target.Path = strings.TrimRight(base.Path, "/") + "/readyz"
		req, _ := http.NewRequestWithContext(r.Context(), http.MethodGet, target.String(), nil)
		response, err := h.client.Do(req)
		if err != nil || response.StatusCode >= 300 {
			failures = append(failures, name)
		}
		if response != nil {
			response.Body.Close()
		}
	}
	if len(failures) > 0 {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "not ready", "services": failures})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (h *Handler) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Origin") == h.allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", h.allowedOrigin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Customer-ID, traceparent")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func copyHeader(dst, src http.Header, key string) {
	if value := src.Get(key); value != "" {
		dst.Set(key, value)
	}
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func Check(ctx context.Context, client *http.Client, target string) bool {
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	response, err := client.Do(request)
	if err != nil {
		return false
	}
	defer response.Body.Close()
	return response.StatusCode < 300
}
