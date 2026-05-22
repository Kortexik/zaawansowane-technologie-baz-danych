package benchrunner

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLRunner runs benchmarks for MySQL
type MySQLRunner struct {
	db     *sql.DB
	config MySQLConfig
}

// NewMySQLRunner creates a new MySQL benchmark runner
func NewMySQLRunner(config MySQLConfig) *MySQLRunner {
	return &MySQLRunner{
		config: config,
	}
}

// Connect establishes connection to MySQL
func (r *MySQLRunner) Connect() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		r.config.User,
		r.config.Password,
		r.config.Host,
		r.config.Port,
		r.config.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	r.db = db
	fmt.Println("✓ Connected to MySQL")
	return nil
}

// Close closes the database connection
func (r *MySQLRunner) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// Explain runs `EXPLAIN <query>` and returns the result as a human-readable
// pipe-separated table. Returns the unparsed plan so the report can include
// it verbatim.
func (r *MySQLRunner) Explain(query string) (string, error) {
	rows, err := r.db.Query("EXPLAIN " + query)
	if err != nil {
		return "", fmt.Errorf("explain: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("columns: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(strings.Join(cols, " | "))
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("-", len(strings.Join(cols, " | "))))
	sb.WriteString("\n")

	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return "", fmt.Errorf("scan: %w", err)
		}
		strs := make([]string, len(cols))
		for i, v := range vals {
			switch t := v.(type) {
			case nil:
				strs[i] = "NULL"
			case []byte:
				strs[i] = string(t)
			default:
				strs[i] = fmt.Sprintf("%v", t)
			}
		}
		sb.WriteString(strings.Join(strs, " | "))
		sb.WriteString("\n")
	}
	return sb.String(), rows.Err()
}

// Reset truncates all benchmark tables so a fresh data load can begin.
func (r *MySQLRunner) Reset() error {
	tables := []string{
		"order_items", "payments", "reviews", "inventory_logs",
		"product_images", "orders", "products", "addresses",
		"categories", "users",
	}
	if _, err := r.db.Exec("SET FOREIGN_KEY_CHECKS=0"); err != nil {
		return fmt.Errorf("disable FK checks: %w", err)
	}
	for _, t := range tables {
		if _, err := r.db.Exec("TRUNCATE TABLE " + t); err != nil {
			fmt.Printf("  Truncate notice (%s): %v\n", t, err)
		}
	}
	if _, err := r.db.Exec("SET FOREIGN_KEY_CHECKS=1"); err != nil {
		return fmt.Errorf("re-enable FK checks: %w", err)
	}
	fmt.Println("✓ MySQL tables truncated")
	return nil
}

// LoadInsertsFromFile streams the first `count` INSERT statements from
// `queryFile` and executes them in 5k-row transactions. Streaming + tx
// batching keeps memory flat for 10M-row files and gives ~10× the throughput
// of auto-commit single-row inserts.
func (r *MySQLRunner) LoadInsertsFromFile(queryFile string, count int) error {
	const txBatch = 5000
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	rowsInTx := 0
	loaded := 0
	_, err = streamSQLStatements(queryFile, count, func(query string) error {
		if strings.HasPrefix(strings.ToUpper(query), "INSERT INTO") {
			query = "INSERT IGNORE" + query[6:]
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

// CreateIndexes creates indexes on non-index test columns to measure index impact
func (r *MySQLRunner) CreateIndexes() error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_country ON users(country)",
		"CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id)",
		"CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)",
		"CREATE INDEX IF NOT EXISTS idx_categories_display_order ON categories(display_order)",
		"CREATE INDEX IF NOT EXISTS idx_addresses_postal_code ON addresses(postal_code)",
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
func (r *MySQLRunner) DropIndexes() error {
	indexes := []string{
		"ALTER TABLE users DROP INDEX IF EXISTS idx_users_country",
		"ALTER TABLE products DROP INDEX IF EXISTS idx_products_category_id",
		"ALTER TABLE orders DROP INDEX IF EXISTS idx_orders_status",
		"ALTER TABLE categories DROP INDEX IF EXISTS idx_categories_display_order",
		"ALTER TABLE addresses DROP INDEX IF EXISTS idx_addresses_postal_code",
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

// RunQueryFile streams up to `limit` queries from the file, then cycles them
// (if fewer were read) to fill exactly `limit` executions for timing.
func (r *MySQLRunner) RunQueryFile(queryFile string, limit int) (*Result, error) {
	rawQueries := make([]string, 0, limit)
	if _, err := streamSQLStatements(queryFile, limit, func(query string) error {
		if strings.HasPrefix(strings.ToUpper(query), "INSERT INTO") {
			query = "INSERT IGNORE" + query[6:]
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
func (r *MySQLRunner) BenchmarkScenario(entity, operation, queryFile string, batchSize, numTrials int) ([]*Result, error) {
	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Printf("MySQL Benchmark: %s - %s (batch: %d)\n", entity, strings.ToUpper(operation), batchSize)
	fmt.Printf("%s\n", strings.Repeat("=", 60))

	results := []*Result{}

	for trial := 1; trial <= numTrials; trial++ {
		fmt.Printf("\nTrial %d/%d:\n", trial, numTrials)

		result, err := r.RunQueryFile(queryFile, batchSize)
		if err != nil {
			return nil, fmt.Errorf("trial %d failed: %w", trial, err)
		}

		result.Database = "mysql"
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

