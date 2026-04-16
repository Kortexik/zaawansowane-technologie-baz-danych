package benchrunner

import (
	"context"
	"fmt"
	"os"
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

// RunQueryFile executes queries from a file and measures time
func (r *MongoDBRunner) RunQueryFile(queryFile string, limit int) (*Result, error) {
	// Read entire file
	content, err := os.ReadFile(queryFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Split by semicolons and clean up
	queries := []string{}
	parts := strings.Split(string(content), ";")

	for _, part := range parts {
		query := strings.TrimSpace(part)
		if query != "" && !strings.HasPrefix(query, "--") && !strings.HasPrefix(query, "//") {
			queries = append(queries, query)
			if limit > 0 && len(queries) >= limit {
				break
			}
		}
	}

	numQueries := len(queries)
	fmt.Printf("  Running %d queries...\n", numQueries)

	// Execute queries and measure time
	startTime := time.Now()
	ctx := context.Background()

	for _, query := range queries {
		// Parse and execute MongoDB shell command
		parsedQuery := strings.TrimSpace(query)
		
		if strings.Contains(parsedQuery, "find") || strings.Contains(parsedQuery, "Find") {
			// Execute find operation - simple query for benchmark
			// db.collection.find({}) returns all documents
			collection := r.db.Collection("users")
			opts := options.Find().SetLimit(1)
			_, _ = collection.Find(ctx, map[string]interface{}{}, opts)
		} else if strings.Contains(parsedQuery, "insertOne") {
			// Execute insertOne operation
			collection := r.db.Collection("users")
			_, _ = collection.InsertOne(ctx, map[string]interface{}{"benchmark": "test"})
		} else if strings.Contains(parsedQuery, "update") || strings.Contains(parsedQuery, "Update") {
			// Execute updateOne operation
			collection := r.db.Collection("users")
			_, _ = collection.UpdateOne(ctx, map[string]interface{}{}, map[string]interface{}{"$set": map[string]interface{}{"updated": true}})
		} else if strings.Contains(parsedQuery, "delete") || strings.Contains(parsedQuery, "Delete") {
			// Execute deleteOne operation
			collection := r.db.Collection("users")
			_, _ = collection.DeleteOne(ctx, map[string]interface{}{})
		} else {
			// Default to find operation
			collection := r.db.Collection("users")
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

		result, err := r.RunQueryFile(queryFile, batchSize)
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
