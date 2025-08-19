package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/exotic-travel-booking/backend/internal/metrics"
)

// MetricsHandlers handles metrics-related HTTP requests
type MetricsHandlers struct {
	collector *metrics.MetricsCollector
}

// NewMetricsHandlers creates a new metrics handlers instance
func NewMetricsHandlers() *MetricsHandlers {
	return &MetricsHandlers{
		collector: metrics.GetGlobalCollector(),
	}
}

// GetMetrics returns current performance metrics
func (h *MetricsHandlers) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if h.collector == nil {
		http.Error(w, "Metrics collector not available", http.StatusServiceUnavailable)
		return
	}

	snapshot := h.collector.GetMetrics()
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	
	if err := json.NewEncoder(w).Encode(snapshot); err != nil {
		http.Error(w, "Failed to encode metrics", http.StatusInternalServerError)
		return
	}
}

// GetHealthMetrics returns health-focused metrics
func (h *MetricsHandlers) GetHealthMetrics(w http.ResponseWriter, r *http.Request) {
	if h.collector == nil {
		http.Error(w, "Metrics collector not available", http.StatusServiceUnavailable)
		return
	}

	snapshot := h.collector.GetMetrics()
	
	// Calculate health indicators
	healthMetrics := map[string]interface{}{
		"status": "healthy",
		"timestamp": snapshot.Timestamp,
		"uptime_seconds": snapshot.Timestamp.Unix(),
		"memory_usage_mb": float64(snapshot.MemoryUsage) / 1024 / 1024,
		"goroutine_count": snapshot.GoroutineCount,
		"request_count": snapshot.RequestCount,
		"error_count": snapshot.ErrorCount,
		"error_rate": calculateErrorRate(snapshot.ErrorCount, snapshot.RequestCount),
		"db_queries": snapshot.DBQueries,
		"db_errors": snapshot.DBErrors,
		"db_error_rate": calculateErrorRate(snapshot.DBErrors, snapshot.DBQueries),
		"cache_hits": snapshot.CacheHits,
		"cache_misses": snapshot.CacheMisses,
		"cache_hit_rate": calculateHitRate(snapshot.CacheHits, snapshot.CacheMisses),
	}

	// Determine overall health status
	if snapshot.ErrorCount > 0 && calculateErrorRate(snapshot.ErrorCount, snapshot.RequestCount) > 5.0 {
		healthMetrics["status"] = "degraded"
	}
	
	if snapshot.GoroutineCount > 10000 {
		healthMetrics["status"] = "warning"
		healthMetrics["warning"] = "High goroutine count detected"
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	
	if err := json.NewEncoder(w).Encode(healthMetrics); err != nil {
		http.Error(w, "Failed to encode health metrics", http.StatusInternalServerError)
		return
	}
}

// GetPrometheusMetrics returns metrics in Prometheus format
func (h *MetricsHandlers) GetPrometheusMetrics(w http.ResponseWriter, r *http.Request) {
	if h.collector == nil {
		http.Error(w, "Metrics collector not available", http.StatusServiceUnavailable)
		return
	}

	snapshot := h.collector.GetMetrics()
	
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	
	// Write Prometheus format metrics
	writePrometheusMetric(w, "http_requests_total", "counter", "Total number of HTTP requests", float64(snapshot.RequestCount))
	writePrometheusMetric(w, "http_errors_total", "counter", "Total number of HTTP errors", float64(snapshot.ErrorCount))
	writePrometheusMetric(w, "http_request_duration_seconds", "histogram", "HTTP request duration", float64(snapshot.RequestDuration.Seconds()))
	
	writePrometheusMetric(w, "db_queries_total", "counter", "Total number of database queries", float64(snapshot.DBQueries))
	writePrometheusMetric(w, "db_errors_total", "counter", "Total number of database errors", float64(snapshot.DBErrors))
	writePrometheusMetric(w, "db_query_duration_seconds", "histogram", "Database query duration", float64(snapshot.DBQueryDuration.Seconds()))
	
	writePrometheusMetric(w, "cache_hits_total", "counter", "Total number of cache hits", float64(snapshot.CacheHits))
	writePrometheusMetric(w, "cache_misses_total", "counter", "Total number of cache misses", float64(snapshot.CacheMisses))
	writePrometheusMetric(w, "cache_errors_total", "counter", "Total number of cache errors", float64(snapshot.CacheErrors))
	
	writePrometheusMetric(w, "memory_usage_bytes", "gauge", "Current memory usage in bytes", float64(snapshot.MemoryUsage))
	writePrometheusMetric(w, "goroutines_count", "gauge", "Current number of goroutines", float64(snapshot.GoroutineCount))
	writePrometheusMetric(w, "gc_pause_total_seconds", "counter", "Total GC pause time", float64(snapshot.GCPauses.Seconds()))
	
	// Custom counters
	for name, value := range snapshot.CustomCounters {
		writePrometheusMetric(w, "custom_"+name+"_total", "counter", "Custom counter: "+name, float64(value))
	}
	
	// Custom gauges
	for name, value := range snapshot.CustomGauges {
		writePrometheusMetric(w, "custom_"+name, "gauge", "Custom gauge: "+name, value)
	}
	
	// Status codes
	for code, count := range snapshot.StatusCodes {
		writePrometheusMetricWithLabels(w, "http_responses_total", "counter", "HTTP responses by status code", 
			map[string]string{"status_code": strconv.Itoa(code)}, float64(count))
	}
}

// RecordCustomMetric allows external recording of custom metrics
func (h *MetricsHandlers) RecordCustomMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.collector == nil {
		http.Error(w, "Metrics collector not available", http.StatusServiceUnavailable)
		return
	}

	var request struct {
		Name  string  `json:"name"`
		Value float64 `json:"value"`
		Type  string  `json:"type"` // "counter", "gauge", "histogram"
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if request.Name == "" {
		http.Error(w, "Metric name is required", http.StatusBadRequest)
		return
	}

	switch request.Type {
	case "counter":
		h.collector.IncrementCounter(request.Name, int64(request.Value))
	case "gauge":
		h.collector.SetGauge(request.Name, request.Value)
	case "histogram":
		h.collector.ObserveHistogram(request.Name, request.Value)
	default:
		http.Error(w, "Invalid metric type. Use 'counter', 'gauge', or 'histogram'", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status": "recorded",
		"name":   request.Name,
		"type":   request.Type,
	}
	
	json.NewEncoder(w).Encode(response)
}

// Helper functions

func calculateErrorRate(errors, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(errors) / float64(total) * 100
}

func calculateHitRate(hits, misses int64) float64 {
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}

func writePrometheusMetric(w http.ResponseWriter, name, metricType, help string, value float64) {
	w.Write([]byte("# HELP " + name + " " + help + "\n"))
	w.Write([]byte("# TYPE " + name + " " + metricType + "\n"))
	w.Write([]byte(name + " " + strconv.FormatFloat(value, 'f', -1, 64) + "\n"))
}

func writePrometheusMetricWithLabels(w http.ResponseWriter, name, metricType, help string, labels map[string]string, value float64) {
	w.Write([]byte("# HELP " + name + " " + help + "\n"))
	w.Write([]byte("# TYPE " + name + " " + metricType + "\n"))
	
	labelStr := ""
	if len(labels) > 0 {
		labelPairs := make([]string, 0, len(labels))
		for k, v := range labels {
			labelPairs = append(labelPairs, k+"=\""+v+"\"")
		}
		labelStr = "{" + joinStrings(labelPairs, ",") + "}"
	}
	
	w.Write([]byte(name + labelStr + " " + strconv.FormatFloat(value, 'f', -1, 64) + "\n"))
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
