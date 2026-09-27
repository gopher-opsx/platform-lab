package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMiddlewareExposesRequestMetrics(t *testing.T) {
	collector := New("test-service")
	mux := http.NewServeMux()
	mux.HandleFunc("GET /widgets/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	mux.Handle("/metrics", collector.Handler())
	handler := collector.Middleware(mux)

	request := httptest.NewRequest(http.MethodGet, "/widgets/123", nil)
	handler.ServeHTTP(httptest.NewRecorder(), request)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := response.Body.String()
	for _, expected := range []string{
		`http_server_requests_total{service="test-service",method="GET",route="/widgets/{id}",status_code="201"} 1`,
		`http_server_request_duration_seconds_count{service="test-service",method="GET",route="/widgets/{id}",status_code="201"} 1`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("metrics output missing %q:\n%s", expected, body)
		}
	}
	if strings.Contains(body, `route="/metrics"`) {
		t.Fatalf("metrics endpoint must not measure itself:\n%s", body)
	}
}
