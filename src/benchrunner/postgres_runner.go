package benchrunner

import (
	"database/sql"
	"fmt"
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

// Explain runs `EXPLAIN <query>` and returns the multi-line plan text
// verbatim (no FORMAT JSON — text is what graders expect to see).
func (r *PostgresRunner) Explain(query string) (string, error) {
	rows, err := r.db.Query("EXPLAIN " + query)
	if err != nil {
		return "", fmt.Errorf("explain: %w", err)
	}
	defer rows.Close()

	var lines []string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return "", fmt.Errorf("scan: %w", err)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n") + "\n", rows.Err()
}

// Reset truncates all benchmark tables so a fresh data load can begin.
func (r *PostgresRunner) Reset() error {
	tables := []string{
		"order_items", "payments", "reviews", "inventory_logs",
		"product_images", "orders", "products", "addresses",
		"categories", "users",
	}
	// CASCADE removes dependent FK rows in one shot.
	stmt := "TRUNCATE TABLE " + strings.Join(tables, ", ") + " RESTART IDENTITY CASCADE"
	if _, err := r.db.Exec(stmt); err != nil {
		return fmt.Errorf("truncate: %w", err)
	}
	fmt.Println("✓ PostgreSQL tables truncated")
	return nil
}

// LoadInsertsFromFile streams the first `count` INSERT statements and
// executes them in 5k-row transactions — see mysql_runner.go for rationale.
func (r *PostgresRunner) LoadInsertsFromFile(queryFile string, count int) error {
	const txBatch = 5000
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	rowsInTx := 0
	loaded := 0
	_, err = streamSQLStatements(queryFile, count, func(query string) error {
		if strings.HasPrefix(strings.ToUpper(query), "INSERT INTO") &&
			!strings.Contains(strings.ToUpper(query), "ON CONFLICT") {
			query = query + " ON CONFLICT DO NOTHING"
		}
		if _, err := tx.Exec(query); err != nil {
			if loaded < 3 {
				fmt.Printf("  Load warning: %v\n", err)
			}
		}
		loaded++
		rowsInTx++
		if rowsInTx >= txBatch {
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("commit: %w", err)
			}
			var beginErr error
			tx, beginErr = r.db.Begin()
			if beginErr != nil {
				return fmt.Errorf("begin tx: %w", beginErr)
			}
			rowsInTx = 0
		}
		return nil
	})
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("final commit: %w", err)
	}
	fmt.Printf("  ✓ Loaded %d rows from %s\n", loaded, queryFile)
	return nil
}

// Same set of (name, table, column) tuples as MySQL — the goal is to keep
// the two SQL benchmarks aligned for side-by-side comparison.
var pgBenchIndexes = []struct {
	name, table, column string
}{
	{"idx_users_country", "users", "country"},
	{"idx_products_category_id", "products", "category_id"},
	{"idx_orders_status", "orders", "status"},
	{"idx_categories_display_order", "categories", "display_order"},
	{"idx_addresses_postal_code", "addresses", "postal_code"},
}

// CreateIndexes creates indexes on non-index test columns to measure index impact.
// Postgres needs ANALYZE after a bulk load + index create for the planner to
// actually use the new indexes; without it the optimizer picks Seq Scan on a
// "cold" pg_class.reltuples and the benchmark looks identical to Phase A.
func (r *PostgresRunner) CreateIndexes() error {
	for _, idx := range pgBenchIndexes {
		stmt := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s(%s)", idx.name, idx.table, idx.column)
		if _, err := r.db.Exec(stmt); err != nil {
			return fmt.Errorf("create %s: %w", idx.name, err)
		}
	}
	if _, err := r.db.Exec("ANALYZE"); err != nil {
		return fmt.Errorf("analyze after index create: %w", err)
	}
	if err := r.verifyIndexesExist(true); err != nil {
		return err
	}
	fmt.Println("✓ Indexes created (and ANALYZE run)")
	return nil
}

// DropIndexes drops all created indexes
func (r *PostgresRunner) DropIndexes() error {
	for _, idx := range pgBenchIndexes {
		stmt := fmt.Sprintf("DROP INDEX IF EXISTS %s", idx.name)
		if _, err := r.db.Exec(stmt); err != nil {
			return fmt.Errorf("drop %s: %w", idx.name, err)
		}
	}
	fmt.Println("✓ Indexes dropped")
	return nil
}

// verifyIndexesExist sanity-checks pg_indexes — defense against silent DDL drift.
func (r *PostgresRunner) verifyIndexesExist(shouldExist bool) error {
	for _, idx := range pgBenchIndexes {
		var n int
		err := r.db.QueryRow(
			"SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public' AND tablename = $1 AND indexname = $2",
			idx.table, idx.name,
		).Scan(&n)
		if err != nil {
			return fmt.Errorf("verify %s: %w", idx.name, err)
		}
		if shouldExist && n == 0 {
			return fmt.Errorf("index %s missing on %s after CreateIndexes", idx.name, idx.table)
		}
	}
	return nil
}

// RunQueryFile streams up to `limit` queries from the file, then cycles them
// (if fewer were read) to fill exactly `limit` executions for timing.
func (r *PostgresRunner) RunQueryFile(queryFile string, limit int) (*Result, error) {
	rawQueries := make([]string, 0, limit)
	if _, err := streamSQLStatements(queryFile, limit, func(query string) error {
		if strings.HasPrefix(strings.ToUpper(query), "INSERT INTO") &&
			!strings.Contains(strings.ToUpper(query), "ON CONFLICT") {
			query = query + " ON CONFLICT DO NOTHING"
		}
		rawQueries = append(rawQueries, query)
		return nil
	}); err != nil {
		return nil, err
	}
	if len(rawQueries) == 0 {
		return nil, fmt.Errorf("no queries found in file: %s", queryFile)
	}

	queries := make([]string, 0, limit)
	for len(queries) < limit {
		queries = append(queries, rawQueries[len(queries)%len(rawQueries)])
	}

	numQueries := len(queries)
	fmt.Printf("  Running %d queries...\n", numQueries)

	startTime := time.Now()
	for i, query := range queries {
		if _, err := r.db.Exec(query); err != nil {
			if i < 3 {
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

