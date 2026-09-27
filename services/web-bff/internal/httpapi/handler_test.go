package httpapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func downstream(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/readyz" {
			_, _ = w.Write([]byte("ready"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"path":"`+r.URL.Path+`","customer":"`+r.Header.Get("X-Customer-ID")+`"}`)
	}))
}

func TestRoutesProductToCatalog(t *testing.T) {
	server := downstream(t)
	defer server.Close()
	h, err := New(server.URL, server.URL, server.URL, "http://localhost:4200")
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodGet, "/api/products/prod-001", nil)
	w := httptest.NewRecorder()
	h.Routes().ServeHTTP(w, r)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"path":"/products/prod-001"`) {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestForwardsCustomerHeaderToOrder(t *testing.T) {
	server := downstream(t)
	defer server.Close()
	h, _ := New(server.URL, server.URL, server.URL, "http://localhost:4200")
	r := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(`{}`))
	r.Header.Set("X-Customer-ID", "customer-1")
	w := httptest.NewRecorder()
	h.Routes().ServeHTTP(w, r)
	if !strings.Contains(w.Body.String(), `"customer":"customer-1"`) {
		t.Fatalf("body=%s", w.Body.String())
	}
}

func TestCORSAllowsConfiguredStorefront(t *testing.T) {
	server := downstream(t)
	defer server.Close()
	h, _ := New(server.URL, server.URL, server.URL, "http://localhost:4200")
	r := httptest.NewRequest(http.MethodOptions, "/api/orders", nil)
	r.Header.Set("Origin", "http://localhost:4200")
	w := httptest.NewRecorder()
	h.Routes().ServeHTTP(w, r)
	if w.Code != http.StatusNoContent || w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:4200" {
		t.Fatalf("code=%d headers=%v", w.Code, w.Header())
	}
}

func TestReadinessChecksAllServices(t *testing.T) {
	server := downstream(t)
	defer server.Close()
	h, _ := New(server.URL, server.URL, server.URL, "")
	r := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	h.Routes().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}
