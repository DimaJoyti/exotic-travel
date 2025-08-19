package metrics

import (
	"runtime"
	"sync"
	"time"
)

// Metrics holds performance metrics
type Metrics struct {
	mu sync.RWMutex

	// HTTP metrics
	RequestCount    int64
	RequestDuration time.Duration
	ErrorCount      int64
	StatusCodes     map[int]int64

	// Database metrics
	DBConnections   int64
	DBQueries       int64
	DBQueryDuration time.Duration
	DBErrors        int64

	// Cache metrics
	CacheHits   int64
	CacheMisses int64
	CacheErrors int64

	// System metrics
	MemoryUsage    uint64
	CPUUsage       float64
	GoroutineCount int
	GCPauses       time.Duration

	// Custom metrics
	CustomCounters   map[string]int64
	CustomGauges     map[string]float64
	CustomHistograms map[string]*Histogram
}

// Histogram tracks distribution of values
type Histogram struct {
	mu      sync.RWMutex
	buckets []float64
	counts  []int64
	sum     float64
	count   int64
}

// NewHistogram creates a new histogram with specified buckets
func NewHistogram(buckets []float64) *Histogram {
	return &Histogram{
		buckets: buckets,
		counts:  make([]int64, len(buckets)+1),
	}
}

// Observe adds a value to the histogram
func (h *Histogram) Observe(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.sum += value
	h.count++

	// Find the appropriate bucket
	for i, bucket := range h.buckets {
		if value <= bucket {
			h.counts[i]++
			return
		}
	}
	// Value is greater than all buckets
	h.counts[len(h.buckets)]++
}

// Summary returns histogram summary
func (h *Histogram) Summary() HistogramSummary {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return HistogramSummary{
		Count:   h.count,
		Sum:     h.sum,
		Buckets: append([]float64{}, h.buckets...),
		Counts:  append([]int64{}, h.counts...),
	}
}

// HistogramSummary contains histogram data
type HistogramSummary struct {
	Count   int64
	Sum     float64
	Buckets []float64
	Counts  []int64
}

// MetricsCollector collects and manages metrics
type MetricsCollector struct {
	metrics *Metrics
	ticker  *time.Ticker
	done    chan struct{}
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: &Metrics{
			StatusCodes:      make(map[int]int64),
			CustomCounters:   make(map[string]int64),
			CustomGauges:     make(map[string]float64),
			CustomHistograms: make(map[string]*Histogram),
		},
		done: make(chan struct{}),
	}
}

// Start begins collecting system metrics
func (mc *MetricsCollector) Start(interval time.Duration) {
	mc.ticker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-mc.ticker.C:
				mc.collectSystemMetrics()
			case <-mc.done:
				return
			}
		}
	}()
}

// Stop stops the metrics collector
func (mc *MetricsCollector) Stop() {
	if mc.ticker != nil {
		mc.ticker.Stop()
	}
	close(mc.done)
}

// collectSystemMetrics collects system-level metrics
func (mc *MetricsCollector) collectSystemMetrics() {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	// Memory metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	mc.metrics.MemoryUsage = m.Alloc

	// Goroutine count
	mc.metrics.GoroutineCount = runtime.NumGoroutine()

	// GC metrics
	mc.metrics.GCPauses = time.Duration(m.PauseTotalNs)
}

// RecordHTTPRequest records an HTTP request metric
func (mc *MetricsCollector) RecordHTTPRequest(duration time.Duration, statusCode int, isError bool) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.RequestCount++
	mc.metrics.RequestDuration += duration
	mc.metrics.StatusCodes[statusCode]++

	if isError {
		mc.metrics.ErrorCount++
	}
}

// RecordDBQuery records a database query metric
func (mc *MetricsCollector) RecordDBQuery(duration time.Duration, isError bool) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.DBQueries++
	mc.metrics.DBQueryDuration += duration

	if isError {
		mc.metrics.DBErrors++
	}
}

// RecordCacheOperation records a cache operation metric
func (mc *MetricsCollector) RecordCacheOperation(hit bool, isError bool) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	if hit {
		mc.metrics.CacheHits++
	} else {
		mc.metrics.CacheMisses++
	}

	if isError {
		mc.metrics.CacheErrors++
	}
}

// IncrementCounter increments a custom counter
func (mc *MetricsCollector) IncrementCounter(name string, value int64) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.CustomCounters[name] += value
}

// SetGauge sets a custom gauge value
func (mc *MetricsCollector) SetGauge(name string, value float64) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.CustomGauges[name] = value
}

// ObserveHistogram adds a value to a custom histogram
func (mc *MetricsCollector) ObserveHistogram(name string, value float64) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	if _, exists := mc.metrics.CustomHistograms[name]; !exists {
		// Default buckets for response times (in milliseconds)
		buckets := []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000}
		mc.metrics.CustomHistograms[name] = NewHistogram(buckets)
	}

	mc.metrics.CustomHistograms[name].Observe(value)
}

// GetMetrics returns a copy of current metrics
func (mc *MetricsCollector) GetMetrics() MetricsSnapshot {
	mc.metrics.mu.RLock()
	defer mc.metrics.mu.RUnlock()

	// Copy status codes
	statusCodes := make(map[int]int64)
	for k, v := range mc.metrics.StatusCodes {
		statusCodes[k] = v
	}

	// Copy custom counters
	customCounters := make(map[string]int64)
	for k, v := range mc.metrics.CustomCounters {
		customCounters[k] = v
	}

	// Copy custom gauges
	customGauges := make(map[string]float64)
	for k, v := range mc.metrics.CustomGauges {
		customGauges[k] = v
	}

	// Copy custom histograms
	customHistograms := make(map[string]HistogramSummary)
	for k, v := range mc.metrics.CustomHistograms {
		customHistograms[k] = v.Summary()
	}

	return MetricsSnapshot{
		RequestCount:     mc.metrics.RequestCount,
		RequestDuration:  mc.metrics.RequestDuration,
		ErrorCount:       mc.metrics.ErrorCount,
		StatusCodes:      statusCodes,
		DBConnections:    mc.metrics.DBConnections,
		DBQueries:        mc.metrics.DBQueries,
		DBQueryDuration:  mc.metrics.DBQueryDuration,
		DBErrors:         mc.metrics.DBErrors,
		CacheHits:        mc.metrics.CacheHits,
		CacheMisses:      mc.metrics.CacheMisses,
		CacheErrors:      mc.metrics.CacheErrors,
		MemoryUsage:      mc.metrics.MemoryUsage,
		CPUUsage:         mc.metrics.CPUUsage,
		GoroutineCount:   mc.metrics.GoroutineCount,
		GCPauses:         mc.metrics.GCPauses,
		CustomCounters:   customCounters,
		CustomGauges:     customGauges,
		CustomHistograms: customHistograms,
		Timestamp:        time.Now(),
	}
}

// MetricsSnapshot represents a point-in-time snapshot of metrics
type MetricsSnapshot struct {
	RequestCount     int64
	RequestDuration  time.Duration
	ErrorCount       int64
	StatusCodes      map[int]int64
	DBConnections    int64
	DBQueries        int64
	DBQueryDuration  time.Duration
	DBErrors         int64
	CacheHits        int64
	CacheMisses      int64
	CacheErrors      int64
	MemoryUsage      uint64
	CPUUsage         float64
	GoroutineCount   int
	GCPauses         time.Duration
	CustomCounters   map[string]int64
	CustomGauges     map[string]float64
	CustomHistograms map[string]HistogramSummary
	Timestamp        time.Time
}

// CalculateRates calculates rate-based metrics
func (ms *MetricsSnapshot) CalculateRates(previous *MetricsSnapshot) MetricsRates {
	if previous == nil {
		return MetricsRates{}
	}

	duration := ms.Timestamp.Sub(previous.Timestamp).Seconds()
	if duration <= 0 {
		return MetricsRates{}
	}

	return MetricsRates{
		RequestsPerSecond:  float64(ms.RequestCount-previous.RequestCount) / duration,
		ErrorsPerSecond:    float64(ms.ErrorCount-previous.ErrorCount) / duration,
		DBQueriesPerSecond: float64(ms.DBQueries-previous.DBQueries) / duration,
		CacheHitRate:       calculateHitRate(ms.CacheHits, ms.CacheMisses, previous.CacheHits, previous.CacheMisses),
		AvgResponseTime:    calculateAvgDuration(ms.RequestDuration, ms.RequestCount, previous.RequestDuration, previous.RequestCount),
		AvgDBQueryTime:     calculateAvgDuration(ms.DBQueryDuration, ms.DBQueries, previous.DBQueryDuration, previous.DBQueries),
	}
}

// MetricsRates contains calculated rate metrics
type MetricsRates struct {
	RequestsPerSecond  float64
	ErrorsPerSecond    float64
	DBQueriesPerSecond float64
	CacheHitRate       float64
	AvgResponseTime    time.Duration
	AvgDBQueryTime     time.Duration
}

// calculateHitRate calculates cache hit rate
func calculateHitRate(hits, misses, prevHits, prevMisses int64) float64 {
	totalOps := (hits + misses) - (prevHits + prevMisses)
	if totalOps <= 0 {
		return 0
	}

	hitOps := hits - prevHits
	return float64(hitOps) / float64(totalOps) * 100
}

// calculateAvgDuration calculates average duration
func calculateAvgDuration(totalDuration time.Duration, count int64, prevTotalDuration time.Duration, prevCount int64) time.Duration {
	operations := count - prevCount
	if operations <= 0 {
		return 0
	}

	duration := totalDuration - prevTotalDuration
	return duration / time.Duration(operations)
}

// Global metrics collector instance
var globalCollector *MetricsCollector

// InitGlobalCollector initializes the global metrics collector
func InitGlobalCollector() {
	globalCollector = NewMetricsCollector()
	globalCollector.Start(10 * time.Second) // Collect system metrics every 10 seconds
}

// GetGlobalCollector returns the global metrics collector
func GetGlobalCollector() *MetricsCollector {
	return globalCollector
}

// StopGlobalCollector stops the global metrics collector
func StopGlobalCollector() {
	if globalCollector != nil {
		globalCollector.Stop()
	}
}
