package benchrunner

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// PostgresRunner runs benchmarks for PostgreSQL
type PostgresRunner struct {
	db     *sql.DB
	config PostgresConfig
}

// NewPostgresRunner creates a new PostgreSQL benchmark runner
func NewPostgresRunner(config PostgresConfig) *PostgresRunner {
	return &PostgresRunner{
		config: config,
	}
}

// Connect establishes connection to PostgreSQL
func (r *PostgresRunner) Connect() error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		r.config.Host,
		r.config.Port,
		r.config.User,
		r.config.Password,
		r.config.Database,
		r.config.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	r.db = db
	fmt.Println("✓ Connected to PostgreSQL")
	return nil
}

// Close closes the database connection
func (r *PostgresRunner) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// CreateIndexes creates indexes on non-index test columns to measure index impact
func (r *PostgresRunner) CreateIndexes() error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_country ON users(country)",
		"CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id)",
		"CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)",
		"CREATE INDEX IF NOT EXISTS idx_categories_display_order ON categories(display_order)",
	}

	for _, idx := range indexes {
		if _, err := r.db.Exec(idx); err != nil {
			// Log but continue - index might already exist
			fmt.Printf("  Index creation notice: %v\n", err)
		}
	}
	fmt.Println("✓ Indexes created")
	return nil
}

// DropIndexes drops all created indexes
func (r *PostgresRunner) DropIndexes() error {
	indexes := []string{
		"DROP INDEX IF EXISTS idx_users_country",
		"DROP INDEX IF EXISTS idx_products_category_id",
		"DROP INDEX IF EXISTS idx_orders_status",
		"DROP INDEX IF EXISTS idx_categories_display_order",
	}

	for _, idx := range indexes {
		if _, err := r.db.Exec(idx); err != nil {
			// Log but continue
			fmt.Printf("  Index drop notice: %v\n", err)
		}
	}
	fmt.Println("✓ Indexes dropped")
	return nil
}

// RunQueryFile executes queries from a file and measures time
func (r *PostgresRunner) RunQueryFile(queryFile string, limit int) (*Result, error) {
	// Read entire file (queries are on one line separated by semicolons)
	content, err := os.ReadFile(queryFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Split by semicolons and clean up
	rawQueries := []string{}
	parts := strings.Split(string(content), ";")

	for _, part := range parts {
		query := strings.TrimSpace(part)
		if query != "" && !strings.HasPrefix(query, "--") {
			// Make INSERT idempotent for PostgreSQL using ON CONFLICT DO NOTHING
			upperQuery := strings.ToUpper(query)
			if strings.HasPrefix(upperQuery, "INSERT INTO") && !strings.Contains(upperQuery, "ON CONFLICT") {
				query = query + " ON CONFLICT DO NOTHING"
			}
			rawQueries = append(rawQueries, query)
		}
	}

	if len(rawQueries) == 0 {
		return nil, fmt.Errorf("no queries found in file: %s", queryFile)
	}

	// Cycle through queries to reach the desired limit
	queries := make([]string, 0, limit)
	for len(queries) < limit {
		queries = append(queries, rawQueries[len(queries)%len(rawQueries)])
	}

	numQueries := len(queries)
	fmt.Printf("  Running %d queries...\n", numQueries)

	// Execute queries and measure time
	startTime := time.Now()

	for i, query := range queries {
		if _, err := r.db.Exec(query); err != nil {
			// Log error but continue
			if i < 3 { // Only log first few errors
				fmt.Printf("  Warning: Query %d failed: %v\n", i+1, err)
			}
		}
	}

	totalTime := time.Since(startTime)

	return &Result{
		NumQueries:     numQueries,
		TotalTime:      totalTime,
		TotalTimeMs:    float64(totalTime.Milliseconds()),
		AvgTimePerOp:   totalTime / time.Duration(numQueries),
		AvgTimePerOpMs: float64(totalTime.Milliseconds()) / float64(numQueries),
		QueriesPerSec:  float64(numQueries) / totalTime.Seconds(),
		Timestamp:      time.Now(),
	}, nil
}

// BenchmarkScenario runs a benchmark scenario with multiple trials
func (r *PostgresRunner) BenchmarkScenario(entity, operation, queryFile string, batchSize, numTrials int) ([]*Result, error) {
	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Printf("PostgreSQL Benchmark: %s - %s (batch: %d)\n", entity, strings.ToUpper(operation), batchSize)
	fmt.Printf("%s\n", strings.Repeat("=", 60))

	results := []*Result{}

	for trial := 1; trial <= numTrials; trial++ {
		fmt.Printf("\nTrial %d/%d:\n", trial, numTrials)

		result, err := r.RunQueryFile(queryFile, batchSize)
		if err != nil {
			return nil, fmt.Errorf("trial %d failed: %w", trial, err)
		}

		result.Database = "postgresql"
		result.Entity = entity
		result.Operation = operation
		result.BatchSize = batchSize
		result.Trial = trial

		results = append(results, result)

		fmt.Printf("  ✓ Time: %.2fs (%.2f ms/query)\n", 
			result.TotalTime.Seconds(), 
			result.AvgTimePerOpMs)
		fmt.Printf("  ✓ QPS: %.2f\n", result.QueriesPerSec)
	}

	// Calculate and print average
	if len(results) > 0 {
		avgTime := 0.0
		avgQPS := 0.0
		for _, r := range results {
			avgTime += r.TotalTimeMs
			avgQPS += r.QueriesPerSec
		}
		avgTime /= float64(len(results))
		avgQPS /= float64(len(results))

		fmt.Printf("\n📊 Average Results:\n")
		fmt.Printf("  Time: %.2fs\n", avgTime/1000)
		fmt.Printf("  QPS: %.2f\n", avgQPS)
	}

	return results, nil
}

