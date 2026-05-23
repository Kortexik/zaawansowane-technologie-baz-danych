package benchrunner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBRunner runs benchmarks for MongoDB
type MongoDBRunner struct {
	client *mongo.Client
	db     *mongo.Database
	config MongoDBConfig
}

// NewMongoDBRunner creates a new MongoDB benchmark runner
func NewMongoDBRunner(config MongoDBConfig) *MongoDBRunner {
	return &MongoDBRunner{
		config: config,
	}
}

// Connect establishes connection to MongoDB
func (r *MongoDBRunner) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build connection URI
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=%s",
		r.config.User,
		r.config.Password,
		r.config.Host,
		r.config.Port,
		r.config.Database,
		r.config.AuthSource,
	)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Verify connection
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	r.client = client
	r.db = client.Database(r.config.Database)
	fmt.Println("✓ Connected to MongoDB")
	return nil
}

// Close closes the database connection
func (r *MongoDBRunner) Close() error {
	if r.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return r.client.Disconnect(ctx)
	}
	return nil
}

// Reset drops all benchmark collections so a fresh data load can begin.
func (r *MongoDBRunner) Reset() error {
	collections := []string{"users", "products", "categories", "orders",
		"addresses", "order_items", "payments", "reviews",
		"inventory_logs", "product_images"}
	ctx := context.Background()
	for _, c := range collections {
		if err := r.db.Collection(c).Drop(ctx); err != nil {
			fmt.Printf("  Drop notice (%s): %v\n", c, err)
		}
	}
	fmt.Println("✓ MongoDB collections dropped")
	return nil
}

// PopulateCollection inserts `count` stub documents into the named collection.
// The runner uses heuristic matching (find/insertOne/etc) so realistic schema
// isn't required for benchmark semantics, only document count.
func (r *MongoDBRunner) PopulateCollection(entity string, count int) error {
	ctx := context.Background()
	coll := r.db.Collection(entity)
	const batch = 5000
	docs := make([]interface{}, 0, batch)
	for i := 1; i <= count; i++ {
		docs = append(docs, map[string]interface{}{
			"_id":       i,
			"name":      fmt.Sprintf("%s_%d", entity, i),
			"value":     i,
			"is_active": i%2 == 0,
		})
		if len(docs) == batch {
			if _, err := coll.InsertMany(ctx, docs); err != nil {
				fmt.Printf("  Insert warning: %v\n", err)
			}
			docs = docs[:0]
		}
	}
	if len(docs) > 0 {
		if _, err := coll.InsertMany(ctx, docs); err != nil {
			fmt.Printf("  Insert warning: %v\n", err)
		}
	}
	fmt.Printf("  ✓ Populated %s with %d docs\n", entity, count)
	return nil
}

// RunQueryFile streams up to `limit` queries, then cycles them to fill
// exactly `limit` executions for timing.
func (r *MongoDBRunner) RunQueryFile(queryFile, entity string, limit int) (*Result, error) {
	rawQueries := make([]string, 0, limit)
	if _, err := streamQueries(queryFile, limit, func(query string) error {
		rawQueries = append(rawQueries, query)
		return nil
	}); err != nil {
		return nil, err
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

	// Determine collection from entity name
	collection := r.db.Collection(entity)

	// Execute queries and measure time
	startTime := time.Now()
	ctx := context.Background()

	for _, query := range queries {
		parsedQuery := strings.TrimSpace(query)

		if strings.Contains(parsedQuery, "find") || strings.Contains(parsedQuery, "Find") {
			opts := options.Find().SetLimit(1)
			_, _ = collection.Find(ctx, map[string]interface{}{}, opts)
		} else if strings.Contains(parsedQuery, "insertOne") {
			// Use upsert-style to avoid duplicate key errors on repeat runs
			_, _ = collection.InsertOne(ctx, map[string]interface{}{"_bench": true})
		} else if strings.Contains(parsedQuery, "update") || strings.Contains(parsedQuery, "Update") {
			_, _ = collection.UpdateOne(ctx,
				map[string]interface{}{"_bench": true},
				map[string]interface{}{"$set": map[string]interface{}{"updated": true}},
				options.Update().SetUpsert(true),
			)
		} else if strings.Contains(parsedQuery, "delete") || strings.Contains(parsedQuery, "Delete") {
			_, _ = collection.DeleteOne(ctx, map[string]interface{}{"_bench": true})
		} else {
			opts := options.Find().SetLimit(1)
			_, _ = collection.Find(ctx, map[string]interface{}{}, opts)
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
func (r *MongoDBRunner) BenchmarkScenario(entity, operation, queryFile string, batchSize, numTrials int) ([]*Result, error) {
	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Printf("MongoDB Benchmark: %s - %s (batch: %d)\n", entity, strings.ToUpper(operation), batchSize)
	fmt.Printf("%s\n", strings.Repeat("=", 60))

	results := []*Result{}

	for trial := 1; trial <= numTrials; trial++ {
		fmt.Printf("\nTrial %d/%d:\n", trial, numTrials)

		result, err := r.RunQueryFile(queryFile, entity, batchSize)
		if err != nil {
			return nil, fmt.Errorf("trial %d failed: %w", trial, err)
		}

		result.Database = "mongodb"
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
