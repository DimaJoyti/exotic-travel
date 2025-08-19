package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// Config holds database configuration
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// Pool wraps sql.DB with additional functionality
type Pool struct {
	*sql.DB
	config Config
}

// NewPool creates a new database connection pool with optimized settings
func NewPool(config Config) (*Pool, error) {
	// Set default values if not provided
	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = 25 // Maximum number of open connections
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 5 // Maximum number of idle connections
	}
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = time.Hour // Maximum lifetime of a connection
	}
	if config.ConnMaxIdleTime == 0 {
		config.ConnMaxIdleTime = time.Minute * 15 // Maximum idle time for a connection
	}
	if config.SSLMode == "" {
		config.SSLMode = "disable"
	}

	// Build connection string
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	// Open database connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Database connection pool established with %d max connections", config.MaxOpenConns)

	return &Pool{
		DB:     db,
		config: config,
	}, nil
}

// Stats returns database connection pool statistics
func (p *Pool) Stats() sql.DBStats {
	return p.DB.Stats()
}

// HealthCheck performs a health check on the database
func (p *Pool) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := p.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Check if we can execute a simple query
	var result int
	if err := p.QueryRowContext(ctx, "SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("database query health check failed: %w", err)
	}

	return nil
}

// Close closes the database connection pool
func (p *Pool) Close() error {
	log.Println("Closing database connection pool")
	return p.DB.Close()
}

// Transaction executes a function within a database transaction
func (p *Pool) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := p.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// PreparedStatement represents a prepared statement with caching
type PreparedStatement struct {
	stmt *sql.Stmt
	pool *Pool
}

// PreparedStatementCache caches prepared statements
type PreparedStatementCache struct {
	statements map[string]*PreparedStatement
	pool       *Pool
}

// NewPreparedStatementCache creates a new prepared statement cache
func NewPreparedStatementCache(pool *Pool) *PreparedStatementCache {
	return &PreparedStatementCache{
		statements: make(map[string]*PreparedStatement),
		pool:       pool,
	}
}

// Get returns a prepared statement from cache or creates a new one
func (psc *PreparedStatementCache) Get(ctx context.Context, query string) (*PreparedStatement, error) {
	if stmt, exists := psc.statements[query]; exists {
		return stmt, nil
	}

	sqlStmt, err := psc.pool.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	stmt := &PreparedStatement{
		stmt: sqlStmt,
		pool: psc.pool,
	}

	psc.statements[query] = stmt
	return stmt, nil
}

// Close closes all prepared statements in the cache
func (psc *PreparedStatementCache) Close() error {
	for query, stmt := range psc.statements {
		if err := stmt.stmt.Close(); err != nil {
			log.Printf("Error closing prepared statement for query %s: %v", query, err)
		}
	}
	psc.statements = make(map[string]*PreparedStatement)
	return nil
}

// QueryContext executes a query with the prepared statement
func (ps *PreparedStatement) QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error) {
	return ps.stmt.QueryContext(ctx, args...)
}

// QueryRowContext executes a query that returns a single row
func (ps *PreparedStatement) QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row {
	return ps.stmt.QueryRowContext(ctx, args...)
}

// ExecContext executes a statement that doesn't return rows
func (ps *PreparedStatement) ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error) {
	return ps.stmt.ExecContext(ctx, args...)
}

// QueryBuilder helps build optimized queries
type QueryBuilder struct {
	query  string
	args   []interface{}
	argNum int
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		args:   make([]interface{}, 0),
		argNum: 1,
	}
}

// Select starts a SELECT query
func (qb *QueryBuilder) Select(columns string) *QueryBuilder {
	qb.query = "SELECT " + columns
	return qb
}

// From adds FROM clause
func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.query += " FROM " + table
	return qb
}

// Where adds WHERE clause
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	if len(args) > 0 {
		qb.query += " WHERE " + qb.replacePlaceholders(condition, len(args))
		qb.args = append(qb.args, args...)
	} else {
		qb.query += " WHERE " + condition
	}
	return qb
}

// And adds AND condition
func (qb *QueryBuilder) And(condition string, args ...interface{}) *QueryBuilder {
	if len(args) > 0 {
		qb.query += " AND " + qb.replacePlaceholders(condition, len(args))
		qb.args = append(qb.args, args...)
	} else {
		qb.query += " AND " + condition
	}
	return qb
}

// OrderBy adds ORDER BY clause
func (qb *QueryBuilder) OrderBy(column, direction string) *QueryBuilder {
	qb.query += " ORDER BY " + column + " " + direction
	return qb
}

// Limit adds LIMIT clause
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.query += fmt.Sprintf(" LIMIT %d", limit)
	return qb
}

// Offset adds OFFSET clause
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.query += fmt.Sprintf(" OFFSET %d", offset)
	return qb
}

// Build returns the final query and arguments
func (qb *QueryBuilder) Build() (string, []interface{}) {
	return qb.query, qb.args
}

// replacePlaceholders replaces ? with $1, $2, etc. for PostgreSQL
func (qb *QueryBuilder) replacePlaceholders(condition string, argCount int) string {
	result := condition
	for i := 0; i < argCount; i++ {
		placeholder := fmt.Sprintf("$%d", qb.argNum)
		result = replaceFirst(result, "?", placeholder)
		qb.argNum++
	}
	return result
}

// replaceFirst replaces the first occurrence of old with new
func replaceFirst(s, old, new string) string {
	if idx := findFirst(s, old); idx != -1 {
		return s[:idx] + new + s[idx+len(old):]
	}
	return s
}

// findFirst finds the first occurrence of substr in s
func findFirst(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
