package metrics

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var buckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5}

type key struct {
	method string
	route  string
	status int
}

type observation struct {
	count   uint64
	sum     float64
	buckets []uint64
}

type Collector struct {
	service string
	mu      sync.RWMutex
	values  map[key]*observation
}

func New(service string) *Collector {
	return &Collector{service: service, values: make(map[key]*observation)}
}

func (c *Collector) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(recorder, r)
		route := r.Pattern
		if route == "" {
			route = "unmatched"
		} else if index := strings.IndexByte(route, ' '); index >= 0 {
			route = route[index+1:]
		}
		c.observe(key{method: r.Method, route: route, status: recorder.status}, time.Since(start).Seconds())
	})
}

func (c *Collector) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		c.write(w)
	})
}

func (c *Collector) observe(k key, seconds float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value := c.values[k]
	if value == nil {
		value = &observation{buckets: make([]uint64, len(buckets))}
		c.values[k] = value
	}
	value.count++
	value.sum += seconds
	for index, upper := range buckets {
		if seconds <= upper {
			value.buckets[index]++
		}
	}
}

func (c *Collector) write(w io.Writer) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]key, 0, len(c.values))
	for k := range c.values {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].route != keys[j].route {
			return keys[i].route < keys[j].route
		}
		if keys[i].method != keys[j].method {
			return keys[i].method < keys[j].method
		}
		return keys[i].status < keys[j].status
	})
	fmt.Fprintln(w, "# HELP http_server_requests_total Total HTTP requests received.")
	fmt.Fprintln(w, "# TYPE http_server_requests_total counter")
	for _, k := range keys {
		value := c.values[k]
		fmt.Fprintf(w, "http_server_requests_total%s %d\n", c.labels(k, ""), value.count)
	}
	fmt.Fprintln(w, "# HELP http_server_request_duration_seconds HTTP server request duration.")
	fmt.Fprintln(w, "# TYPE http_server_request_duration_seconds histogram")
	for _, k := range keys {
		value := c.values[k]
		for index, upper := range buckets {
			fmt.Fprintf(w, "http_server_request_duration_seconds_bucket%s %d\n", c.labels(k, strconv.FormatFloat(upper, 'g', -1, 64)), value.buckets[index])
		}
		fmt.Fprintf(w, "http_server_request_duration_seconds_bucket%s %d\n", c.labels(k, "+Inf"), value.count)
		fmt.Fprintf(w, "http_server_request_duration_seconds_sum%s %.9f\n", c.labels(k, ""), value.sum)
		fmt.Fprintf(w, "http_server_request_duration_seconds_count%s %d\n", c.labels(k, ""), value.count)
	}
}

func (c *Collector) labels(k key, le string) string {
	parts := []string{
		`service="` + escape(c.service) + `"`,
		`method="` + escape(k.method) + `"`,
		`route="` + escape(k.route) + `"`,
		`status_code="` + strconv.Itoa(k.status) + `"`,
	}
	if le != "" {
		parts = append(parts, `le="`+le+`"`)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func escape(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\n", "\\n")
	return strings.ReplaceAll(value, `"`, `\\"`)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
