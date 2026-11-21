package observability

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const prometheusContentType = "text/plain; version=0.0.4; charset=utf-8"

var defaultLatencyBuckets = []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

type httpCounterKey struct {
	method      string
	route       string
	statusClass string
}

type httpHistKey struct {
	method string
	route  string
}

type depErrKey struct {
	dependency string
	operation  string
}

type histogramData struct {
	counts []uint64
	sum    float64
	total  uint64
}

// Registry stores Prometheus-compatible RED, dependency, outbox, and Go runtime metrics
// using only the Go standard library (zero external dependencies, low memory footprint).
type Registry struct {
	service   string
	startedAt time.Time
	buckets   []float64

	mu           sync.RWMutex
	httpRequests map[httpCounterKey]uint64
	httpLatency  map[httpHistKey]*histogramData
	inFlight     int64
	depErrors    map[depErrKey]uint64
	outboxEvents map[string]uint64
}

// NewRegistry creates an isolated metric registry for the given service name.
func NewRegistry(service string) *Registry {
	buckets := make([]float64, len(defaultLatencyBuckets))
	copy(buckets, defaultLatencyBuckets)
	return &Registry{
		service:      service,
		startedAt:    time.Now(),
		buckets:      buckets,
		httpRequests: make(map[httpCounterKey]uint64),
		httpLatency:  make(map[httpHistKey]*histogramData),
		depErrors:    make(map[depErrKey]uint64),
		outboxEvents: make(map[string]uint64),
	}
}

// IncInFlight increments the in-flight HTTP request gauge.
func (r *Registry) IncInFlight() {
	r.mu.Lock()
	r.inFlight++
	r.mu.Unlock()
}

// DecInFlight decrements the in-flight HTTP request gauge.
func (r *Registry) DecInFlight() {
	r.mu.Lock()
	if r.inFlight > 0 {
		r.inFlight--
	}
	r.mu.Unlock()
}

// ObserveHTTPRequest records a completed HTTP request count and duration histogram entry.
func (r *Registry) ObserveHTTPRequest(method, route string, status int, duration time.Duration) {
	statusClass := statusCodeClass(status)
	sec := duration.Seconds()

	r.mu.Lock()
	defer r.mu.Unlock()

	ck := httpCounterKey{method: method, route: route, statusClass: statusClass}
	r.httpRequests[ck]++

	hk := httpHistKey{method: method, route: route}
	hd, ok := r.httpLatency[hk]
	if !ok {
		hd = &histogramData{counts: make([]uint64, len(r.buckets))}
		r.httpLatency[hk] = hd
	}
	hd.sum += sec
	hd.total++
	for i, bound := range r.buckets {
		if sec <= bound {
			hd.counts[i]++
		}
	}
}

// IncDependencyError increments the error counter for an external dependency call.
func (r *Registry) IncDependencyError(dependency, operation string) {
	r.mu.Lock()
	r.depErrors[depErrKey{dependency: dependency, operation: operation}]++
	r.mu.Unlock()
}

// IncOutboxEvent increments the outbox publication counter for a given status (published, failed).
func (r *Registry) IncOutboxEvent(status string) {
	r.mu.Lock()
	r.outboxEvents[status]++
	r.mu.Unlock()
}

// MetricsEnabled returns true when METRICS_ENABLED is set to "true" (used for local Docker scraping).
func MetricsEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("METRICS_ENABLED")), "true")
}

// Handler serves metrics in Prometheus text exposition format 0.0.4 when enabled.
func (r *Registry) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", prometheusContentType)
		w.WriteHeader(http.StatusOK)
		_ = r.WritePrometheus(w)
	})
}

// WritePrometheus writes all collected metrics in standard Prometheus text format 0.0.4.
func (r *Registry) WritePrometheus(w io.Writer) error {
	var b strings.Builder
	svc := escapeLabel(r.service)

	r.mu.RLock()
	inFlight := r.inFlight

	httpKeys := make([]httpCounterKey, 0, len(r.httpRequests))
	for k := range r.httpRequests {
		httpKeys = append(httpKeys, k)
	}
	sort.Slice(httpKeys, func(i, j int) bool {
		if httpKeys[i].route != httpKeys[j].route {
			return httpKeys[i].route < httpKeys[j].route
		}
		if httpKeys[i].method != httpKeys[j].method {
			return httpKeys[i].method < httpKeys[j].method
		}
		return httpKeys[i].statusClass < httpKeys[j].statusClass
	})

	b.WriteString("# HELP buenvivir_http_requests_total Total HTTP requests handled by the service.\n")
	b.WriteString("# TYPE buenvivir_http_requests_total counter\n")
	for _, k := range httpKeys {
		fmt.Fprintf(&b, "buenvivir_http_requests_total{service=\"%s\",method=\"%s\",route=\"%s\",status_class=\"%s\"} %d\n",
			svc, escapeLabel(k.method), escapeLabel(k.route), escapeLabel(k.statusClass), r.httpRequests[k])
	}

	histKeys := make([]httpHistKey, 0, len(r.httpLatency))
	for k := range r.httpLatency {
		histKeys = append(histKeys, k)
	}
	sort.Slice(histKeys, func(i, j int) bool {
		if histKeys[i].route != histKeys[j].route {
			return histKeys[i].route < histKeys[j].route
		}
		return histKeys[i].method < histKeys[j].method
	})

	b.WriteString("# HELP buenvivir_http_request_duration_seconds HTTP request latency distribution in seconds.\n")
	b.WriteString("# TYPE buenvivir_http_request_duration_seconds histogram\n")
	for _, k := range histKeys {
		hd := r.httpLatency[k]
		m := escapeLabel(k.method)
		rt := escapeLabel(k.route)
		for i, bound := range r.buckets {
			fmt.Fprintf(&b, "buenvivir_http_request_duration_seconds_bucket{service=\"%s\",method=\"%s\",route=\"%s\",le=\"%s\"} %d\n",
				svc, m, rt, strconv.FormatFloat(bound, 'f', -1, 64), hd.counts[i])
		}
		fmt.Fprintf(&b, "buenvivir_http_request_duration_seconds_bucket{service=\"%s\",method=\"%s\",route=\"%s\",le=\"+Inf\"} %d\n",
			svc, m, rt, hd.total)
		fmt.Fprintf(&b, "buenvivir_http_request_duration_seconds_sum{service=\"%s\",method=\"%s\",route=\"%s\"} %g\n",
			svc, m, rt, hd.sum)
		fmt.Fprintf(&b, "buenvivir_http_request_duration_seconds_count{service=\"%s\",method=\"%s\",route=\"%s\"} %d\n",
			svc, m, rt, hd.total)
	}

	b.WriteString("# HELP buenvivir_http_requests_in_flight Current number of in-flight HTTP requests.\n")
	b.WriteString("# TYPE buenvivir_http_requests_in_flight gauge\n")
	fmt.Fprintf(&b, "buenvivir_http_requests_in_flight{service=\"%s\"} %d\n", svc, inFlight)

	if len(r.depErrors) > 0 {
		depKeys := make([]depErrKey, 0, len(r.depErrors))
		for k := range r.depErrors {
			depKeys = append(depKeys, k)
		}
		sort.Slice(depKeys, func(i, j int) bool {
			if depKeys[i].dependency != depKeys[j].dependency {
				return depKeys[i].dependency < depKeys[j].dependency
			}
			return depKeys[i].operation < depKeys[j].operation
		})
		b.WriteString("# HELP buenvivir_dependency_errors_total Total errors communicating with external dependencies.\n")
		b.WriteString("# TYPE buenvivir_dependency_errors_total counter\n")
		for _, k := range depKeys {
			fmt.Fprintf(&b, "buenvivir_dependency_errors_total{service=\"%s\",dependency=\"%s\",operation=\"%s\"} %d\n",
				svc, escapeLabel(k.dependency), escapeLabel(k.operation), r.depErrors[k])
		}
	}

	if len(r.outboxEvents) > 0 {
		obKeys := make([]string, 0, len(r.outboxEvents))
		for k := range r.outboxEvents {
			obKeys = append(obKeys, k)
		}
		sort.Strings(obKeys)
		b.WriteString("# HELP buenvivir_outbox_events_total Total outbox events processed by status.\n")
		b.WriteString("# TYPE buenvivir_outbox_events_total counter\n")
		for _, k := range obKeys {
			fmt.Fprintf(&b, "buenvivir_outbox_events_total{service=\"%s\",status=\"%s\"} %d\n",
				svc, escapeLabel(k), r.outboxEvents[k])
		}
	}
	r.mu.RUnlock()

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	uptime := time.Since(r.startedAt).Seconds()

	b.WriteString("# HELP go_goroutines Number of goroutines that currently exist.\n")
	b.WriteString("# TYPE go_goroutines gauge\n")
	fmt.Fprintf(&b, "go_goroutines{service=\"%s\"} %d\n", svc, runtime.NumGoroutine())

	b.WriteString("# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use.\n")
	b.WriteString("# TYPE go_memstats_alloc_bytes gauge\n")
	fmt.Fprintf(&b, "go_memstats_alloc_bytes{service=\"%s\"} %d\n", svc, mem.Alloc)

	b.WriteString("# HELP go_memstats_heap_inuse_bytes Number of heap bytes that are in use.\n")
	b.WriteString("# TYPE go_memstats_heap_inuse_bytes gauge\n")
	fmt.Fprintf(&b, "go_memstats_heap_inuse_bytes{service=\"%s\"} %d\n", svc, mem.HeapInuse)

	b.WriteString("# HELP go_memstats_sys_bytes Number of bytes obtained from system.\n")
	b.WriteString("# TYPE go_memstats_sys_bytes gauge\n")
	fmt.Fprintf(&b, "go_memstats_sys_bytes{service=\"%s\"} %d\n", svc, mem.Sys)

	b.WriteString("# HELP go_gc_completed_total Total number of completed GC cycles.\n")
	b.WriteString("# TYPE go_gc_completed_total counter\n")
	fmt.Fprintf(&b, "go_gc_completed_total{service=\"%s\"} %d\n", svc, mem.NumGC)

	b.WriteString("# HELP process_uptime_seconds Seconds since the service process started.\n")
	b.WriteString("# TYPE process_uptime_seconds gauge\n")
	fmt.Fprintf(&b, "process_uptime_seconds{service=\"%s\"} %.3f\n", svc, uptime)

	_, err := io.WriteString(w, b.String())
	return err
}

func statusCodeClass(status int) string {
	switch {
	case status >= 500:
		return "5xx"
	case status >= 400:
		return "4xx"
	case status >= 300:
		return "3xx"
	case status >= 200:
		return "2xx"
	default:
		return "1xx"
	}
}

func escapeLabel(v string) string {
	v = strings.ReplaceAll(v, "\\", "\\\\")
	v = strings.ReplaceAll(v, "\"", "\\\"")
	v = strings.ReplaceAll(v, "\n", "\\n")
	return v
}
