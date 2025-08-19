package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

// QueryOptimizer provides database query optimization utilities
type QueryOptimizer struct {
	db                *Pool
	slowQueryThreshold time.Duration
	enableExplain     bool
}

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer(db *Pool, slowQueryThreshold time.Duration) *QueryOptimizer {
	return &QueryOptimizer{
		db:                db,
		slowQueryThreshold: slowQueryThreshold,
		enableExplain:     true,
	}
}

// OptimizedQuery represents an optimized database query
type OptimizedQuery struct {
	SQL        string
	Args       []interface{}
	Hints      []string
	IndexHints []string
}

// QueryStats holds query execution statistics
type QueryStats struct {
	Query         string
	Duration      time.Duration
	RowsAffected  int64
	RowsReturned  int64
	ExecutionPlan string
	Timestamp     time.Time
}

// ExecuteWithStats executes a query and collects performance statistics
func (qo *QueryOptimizer) ExecuteWithStats(ctx context.Context, query string, args ...interface{}) (*sql.Rows, *QueryStats, error) {
	start := time.Now()
	
	// Execute the query
	rows, err := qo.db.QueryContext(ctx, query, args...)
	duration := time.Since(start)
	
	stats := &QueryStats{
		Query:     query,
		Duration:  duration,
		Timestamp: start,
	}
	
	if err != nil {
		return nil, stats, err
	}
	
	// Log slow queries
	if duration > qo.slowQueryThreshold {
		log.Printf("SLOW QUERY [%v]: %s", duration, query)
		
		if qo.enableExplain {
			plan, explainErr := qo.explainQuery(ctx, query, args...)
			if explainErr == nil {
				stats.ExecutionPlan = plan
				log.Printf("EXECUTION PLAN: %s", plan)
			}
		}
	}
	
	return rows, stats, nil
}

// ExecWithStats executes a non-query statement and collects statistics
func (qo *QueryOptimizer) ExecWithStats(ctx context.Context, query string, args ...interface{}) (sql.Result, *QueryStats, error) {
	start := time.Now()
	
	result, err := qo.db.ExecContext(ctx, query, args...)
	duration := time.Since(start)
	
	stats := &QueryStats{
		Query:     query,
		Duration:  duration,
		Timestamp: start,
	}
	
	if err != nil {
		return nil, stats, err
	}
	
	if result != nil {
		if rowsAffected, raErr := result.RowsAffected(); raErr == nil {
			stats.RowsAffected = rowsAffected
		}
	}
	
	// Log slow queries
	if duration > qo.slowQueryThreshold {
		log.Printf("SLOW EXEC [%v]: %s", duration, query)
	}
	
	return result, stats, nil
}

// explainQuery gets the execution plan for a query
func (qo *QueryOptimizer) explainQuery(ctx context.Context, query string, args ...interface{}) (string, error) {
	explainQuery := "EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) " + query
	
	row := qo.db.QueryRowContext(ctx, explainQuery, args...)
	var plan string
	err := row.Scan(&plan)
	
	return plan, err
}

// OptimizeSelect optimizes a SELECT query
func (qo *QueryOptimizer) OptimizeSelect(query string) *OptimizedQuery {
	optimized := &OptimizedQuery{
		SQL:        query,
		Hints:      []string{},
		IndexHints: []string{},
	}
	
	// Convert to lowercase for analysis
	lowerQuery := strings.ToLower(query)
	
	// Add index hints for common patterns
	if strings.Contains(lowerQuery, "order by") && !strings.Contains(lowerQuery, "limit") {
		optimized.Hints = append(optimized.Hints, "Consider adding LIMIT clause for large result sets")
	}
	
	if strings.Contains(lowerQuery, "like '%") {
		optimized.Hints = append(optimized.Hints, "Leading wildcard LIKE queries cannot use indexes efficiently")
	}
	
	if strings.Contains(lowerQuery, "or") {
		optimized.Hints = append(optimized.Hints, "OR conditions may prevent index usage, consider UNION")
	}
	
	if strings.Contains(lowerQuery, "select *") {
		optimized.Hints = append(optimized.Hints, "Avoid SELECT *, specify only needed columns")
	}
	
	// Suggest indexes for WHERE clauses
	if strings.Contains(lowerQuery, "where") {
		optimized.IndexHints = append(optimized.IndexHints, "Ensure indexes exist on WHERE clause columns")
	}
	
	if strings.Contains(lowerQuery, "join") {
		optimized.IndexHints = append(optimized.IndexHints, "Ensure indexes exist on JOIN columns")
	}
	
	return optimized
}

// BatchInsert performs optimized batch insert operations
func (qo *QueryOptimizer) BatchInsert(ctx context.Context, table string, columns []string, values [][]interface{}, batchSize int) error {
	if len(values) == 0 {
		return nil
	}
	
	if batchSize <= 0 {
		batchSize = 1000 // Default batch size
	}
	
	// Build the base query
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	
	// Process in batches
	for i := 0; i < len(values); i += batchSize {
		end := i + batchSize
		if end > len(values) {
			end = len(values)
		}
		
		batch := values[i:end]
		
		// Build batch query
		valueClauses := make([]string, len(batch))
		args := make([]interface{}, 0, len(batch)*len(columns))
		
		for j, row := range batch {
			// Update placeholders for this row
			rowPlaceholders := make([]string, len(columns))
			for k := range columns {
				placeholder := fmt.Sprintf("$%d", j*len(columns)+k+1)
				rowPlaceholders[k] = placeholder
			}
			
			valueClauses[j] = "(" + strings.Join(rowPlaceholders, ", ") + ")"
			args = append(args, row...)
		}
		
		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES %s",
			table,
			strings.Join(columns, ", "),
			strings.Join(valueClauses, ", "),
		)
		
		_, _, err := qo.ExecWithStats(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("batch insert failed at batch %d: %w", i/batchSize, err)
		}
	}
	
	return nil
}

// UpsertBatch performs optimized batch upsert operations
func (qo *QueryOptimizer) UpsertBatch(ctx context.Context, table string, columns []string, values [][]interface{}, conflictColumns []string, updateColumns []string, batchSize int) error {
	if len(values) == 0 {
		return nil
	}
	
	if batchSize <= 0 {
		batchSize = 500 // Smaller default for upserts
	}
	
	// Build update clause
	updateClauses := make([]string, len(updateColumns))
	for i, col := range updateColumns {
		updateClauses[i] = fmt.Sprintf("%s = EXCLUDED.%s", col, col)
	}
	
	// Process in batches
	for i := 0; i < len(values); i += batchSize {
		end := i + batchSize
		if end > len(values) {
			end = len(values)
		}
		
		batch := values[i:end]
		
		// Build batch query
		valueClauses := make([]string, len(batch))
		args := make([]interface{}, 0, len(batch)*len(columns))
		
		for j, row := range batch {
			rowPlaceholders := make([]string, len(columns))
			for k := range columns {
				placeholder := fmt.Sprintf("$%d", j*len(columns)+k+1)
				rowPlaceholders[k] = placeholder
			}
			
			valueClauses[j] = "(" + strings.Join(rowPlaceholders, ", ") + ")"
			args = append(args, row...)
		}
		
		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES %s ON CONFLICT (%s) DO UPDATE SET %s",
			table,
			strings.Join(columns, ", "),
			strings.Join(valueClauses, ", "),
			strings.Join(conflictColumns, ", "),
			strings.Join(updateClauses, ", "),
		)
		
		_, _, err := qo.ExecWithStats(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("batch upsert failed at batch %d: %w", i/batchSize, err)
		}
	}
	
	return nil
}

// AnalyzeTable analyzes table statistics for query optimization
func (qo *QueryOptimizer) AnalyzeTable(ctx context.Context, tableName string) error {
	query := fmt.Sprintf("ANALYZE %s", tableName)
	_, _, err := qo.ExecWithStats(ctx, query)
	return err
}

// GetTableStats retrieves table statistics
func (qo *QueryOptimizer) GetTableStats(ctx context.Context, tableName string) (map[string]interface{}, error) {
	query := `
		SELECT 
			schemaname,
			tablename,
			attname,
			n_distinct,
			most_common_vals,
			most_common_freqs,
			histogram_bounds
		FROM pg_stats 
		WHERE tablename = $1`
	
	rows, _, err := qo.ExecuteWithStats(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	stats := make(map[string]interface{})
	for rows.Next() {
		var schemaname, tablename, attname sql.NullString
		var nDistinct sql.NullFloat64
		var mostCommonVals, mostCommonFreqs, histogramBounds sql.NullString
		
		err := rows.Scan(&schemaname, &tablename, &attname, &nDistinct, &mostCommonVals, &mostCommonFreqs, &histogramBounds)
		if err != nil {
			return nil, err
		}
		
		if attname.Valid {
			stats[attname.String] = map[string]interface{}{
				"n_distinct":        nDistinct.Float64,
				"most_common_vals":  mostCommonVals.String,
				"most_common_freqs": mostCommonFreqs.String,
				"histogram_bounds":  histogramBounds.String,
			}
		}
	}
	
	return stats, nil
}

// GetSlowQueries retrieves slow queries from pg_stat_statements
func (qo *QueryOptimizer) GetSlowQueries(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			query,
			calls,
			total_time,
			mean_time,
			rows,
			100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent
		FROM pg_stat_statements 
		ORDER BY mean_time DESC 
		LIMIT $1`
	
	rows, _, err := qo.ExecuteWithStats(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var slowQueries []map[string]interface{}
	for rows.Next() {
		var query sql.NullString
		var calls, totalTime, meanTime, rowCount sql.NullFloat64
		var hitPercent sql.NullFloat64
		
		err := rows.Scan(&query, &calls, &totalTime, &meanTime, &rowCount, &hitPercent)
		if err != nil {
			return nil, err
		}
		
		slowQueries = append(slowQueries, map[string]interface{}{
			"query":       query.String,
			"calls":       calls.Float64,
			"total_time":  totalTime.Float64,
			"mean_time":   meanTime.Float64,
			"rows":        rowCount.Float64,
			"hit_percent": hitPercent.Float64,
		})
	}
	
	return slowQueries, nil
}

// CreateIndex creates an index with optimization hints
func (qo *QueryOptimizer) CreateIndex(ctx context.Context, indexName, tableName string, columns []string, unique bool, concurrent bool) error {
	var query strings.Builder
	
	query.WriteString("CREATE ")
	if unique {
		query.WriteString("UNIQUE ")
	}
	query.WriteString("INDEX ")
	if concurrent {
		query.WriteString("CONCURRENTLY ")
	}
	query.WriteString(fmt.Sprintf("%s ON %s (%s)", indexName, tableName, strings.Join(columns, ", ")))
	
	_, _, err := qo.ExecWithStats(ctx, query.String())
	return err
}

// GetIndexUsage retrieves index usage statistics
func (qo *QueryOptimizer) GetIndexUsage(ctx context.Context, tableName string) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			indexrelname,
			idx_tup_read,
			idx_tup_fetch,
			idx_scan
		FROM pg_stat_user_indexes 
		WHERE relname = $1
		ORDER BY idx_scan DESC`
	
	rows, _, err := qo.ExecuteWithStats(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var indexStats []map[string]interface{}
	for rows.Next() {
		var indexName sql.NullString
		var tupRead, tupFetch, idxScan sql.NullInt64
		
		err := rows.Scan(&indexName, &tupRead, &tupFetch, &idxScan)
		if err != nil {
			return nil, err
		}
		
		indexStats = append(indexStats, map[string]interface{}{
			"index_name":    indexName.String,
			"tuples_read":   tupRead.Int64,
			"tuples_fetch":  tupFetch.Int64,
			"index_scans":   idxScan.Int64,
		})
	}
	
	return indexStats, nil
}
