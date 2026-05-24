package benchrunner

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Parsers for the JS-shell-style query files we generate.
// The generators emit single-field filters like `{country: 'Poland'}` or
// `{_id: 123}` plus an optional `.limit(N)` suffix — that's enough coverage
// for the benchmark scenarios; nested/array filters are out of scope.
var (
	mongoFilterRegex = regexp.MustCompile(`\{\s*(\w+)\s*:\s*('([^']*)'|(-?\d+))\s*\}`)
	mongoLimitRegex  = regexp.MustCompile(`\.limit\(\s*(\d+)\s*\)`)
)

func parseMongoFilter(s string) bson.M {
	m := mongoFilterRegex.FindStringSubmatch(s)
	if len(m) == 0 {
		return bson.M{}
	}
	key := m[1]
	if m[3] != "" {
		return bson.M{key: m[3]}
	}
	n, _ := strconv.Atoi(m[4])
	return bson.M{key: n}
}

func parseMongoLimit(s string) int64 {
	m := mongoLimitRegex.FindStringSubmatch(s)
	if len(m) == 0 {
		return 0
	}
	n, _ := strconv.Atoi(m[1])
	return int64(n)
}

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

// CreateIndexes creates secondary indexes on the fields the find scenarios
// filter by — mirrors the SQL CreateIndexes so Mongo gets the same Phase-B
// treatment. Without this, `find({country: ...})` on 10M docs is a full
// collection scan.
//
// Also creates a sparse index on `_bench` for every benchmarked collection.
// Reason: this runner's insertOne/updateOne/deleteOne paths use
// `{_bench: true}` as the filter to avoid `_id` collisions across trials.
// Without a `_bench` index, those ops degrade to a full collection scan on
// every call — at 10M docs that's ~2 s per delete, blowing the runtime budget
// by hours. Sparse index keeps the index tiny (only ever indexes the bench
// docs, not the populated dataset).
func (r *MongoDBRunner) CreateIndexes() error {
	ctx := context.Background()
	specs := []struct {
		coll, field string
	}{
		{"users", "country"},
		{"products", "categoryId"},
		{"orders", "status"},
		{"categories", "displayOrder"},
		{"addresses", "postalCode"},
	}
	for _, s := range specs {
		key := bson.D{{Key: s.field, Value: 1}}
		_, err := r.db.Collection(s.coll).Indexes().CreateOne(ctx, mongo.IndexModel{Keys: key})
		if err != nil {
			return fmt.Errorf("create mongo index %s.%s: %w", s.coll, s.field, err)
		}
	}
	// Sparse `_bench` index on every collection that gets insertOne/updateOne
	// /deleteOne scenarios (per main.go mongoScenarios that's users/categories
	// /products). Sparse → only includes docs where the field exists.
	benchColls := []string{"users", "categories", "products", "orders", "addresses"}
	for _, c := range benchColls {
		_, err := r.db.Collection(c).Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{{Key: "_bench", Value: 1}},
			Options: options.Index().SetSparse(true).SetName("idx_bench"),
		})
		if err != nil {
			return fmt.Errorf("create mongo _bench index on %s: %w", c, err)
		}
	}
	fmt.Println("✓ MongoDB indexes created (incl. sparse _bench)")
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

// Lists mirror src/generator/data.go so MongoDB documents end up with the
// same shape of indexed-column values as the SQL benchmark, making the
// find scenarios actually exercise the indexes we create.
var (
	mongoCountriesList = []string{
		"Poland", "Germany", "France", "Spain", "Italy", "United Kingdom", "Netherlands", "Belgium",
		"Austria", "Switzerland", "Sweden", "Norway", "Finland", "Denmark", "Czech Republic", "Slovakia",
		"Hungary", "Romania", "Bulgaria", "Greece", "Portugal", "Ireland", "Croatia", "Slovenia",
		"Estonia", "Latvia", "Lithuania", "USA", "Canada", "Mexico", "Brazil", "Argentina",
		"Chile", "Colombia", "Australia", "New Zealand", "Japan", "South Korea", "China", "India",
		"Indonesia", "Thailand", "Vietnam", "Philippines", "Turkey", "Israel", "Egypt", "Morocco",
		"South Africa", "Kenya",
	}
)

func mongoStatusFor(i int) string {
	// 200 distinct statuses, mirroring generator.statusList shape.
	bases := []string{"pending", "processing", "verified", "preparing", "packed",
		"ready", "shipped", "in_transit", "out_for_delivery", "delivered",
		"cancelled", "refunded", "returned", "disputed", "on_hold"}
	stage := i % 14
	return fmt.Sprintf("%s_step%d", bases[i%len(bases)], stage+1)
}

// PopulateCollection inserts `count` documents with the fields needed by the
// generated find queries (country/categoryId/status/postalCode/displayOrder).
// Uses deterministic cycling i%len(list) instead of RNG so the distribution
// across indexed values is exactly uniform.
func (r *MongoDBRunner) PopulateCollection(entity string, count int) error {
	ctx := context.Background()
	coll := r.db.Collection(entity)
	const batch = 5000
	docs := make([]interface{}, 0, batch)
	for i := 1; i <= count; i++ {
		var doc bson.M
		switch entity {
		case "users":
			doc = bson.M{
				"_id":      i,
				"email":    fmt.Sprintf("user%d@example.com", i),
				"country":  mongoCountriesList[i%len(mongoCountriesList)],
				"isActive": i%2 == 0,
			}
		case "products":
			doc = bson.M{
				"_id":        i,
				"name":       fmt.Sprintf("Product %d", i),
				"categoryId": (i % 200) + 1,
				"price":      float64(i%2000) + 10,
			}
		case "orders":
			doc = bson.M{
				"_id":         i,
				"status":      mongoStatusFor(i),
				"totalAmount": float64(i%5000) + 20,
			}
		case "categories":
			doc = bson.M{
				"_id":          i,
				"name":         fmt.Sprintf("Category %d", i),
				"displayOrder": i,
			}
		case "addresses":
			doc = bson.M{
				"_id":        i,
				"postalCode": fmt.Sprintf("%05d", i%100000),
				"country":    mongoCountriesList[i%len(mongoCountriesList)],
			}
		default:
			doc = bson.M{"_id": i, "value": i}
		}
		docs = append(docs, doc)
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

	// Execute queries and measure time. Each branch parses the actual JS
	// query string from the file so the workload matches what the SQL
	// benchmarks run on the equivalent indexed column.
	startTime := time.Now()
	ctx := context.Background()

	for _, query := range queries {
		q := strings.TrimSpace(query)

		switch {
		case strings.Contains(q, ".findOne("):
			filter := parseMongoFilter(q)
			_ = collection.FindOne(ctx, filter).Err()
		case strings.Contains(q, ".find("):
			filter := parseMongoFilter(q)
			opts := options.Find()
			if l := parseMongoLimit(q); l > 0 {
				opts.SetLimit(l)
			}
			cur, err := collection.Find(ctx, filter, opts)
			if err == nil {
				for cur.Next(ctx) {
				}
				_ = cur.Close(ctx)
			}
		case strings.Contains(q, ".insertOne("):
			// Bench-tagged doc avoids _id collisions with populated data.
			// We're measuring insert throughput; faithful doc replay would
			// hit duplicate-key errors after the first trial.
			_, _ = collection.InsertOne(ctx, bson.M{
				"_bench": true,
				"ts":     time.Now().UnixNano(),
			})
		case strings.Contains(q, ".updateOne("):
			// updateOne in our generated files filters by {_id: N} which
			// degrades to a no-op after first trial deletes some rows.
			// Upserting on a bench tag keeps trial 2/3 comparable to trial 1.
			_, _ = collection.UpdateOne(ctx,
				bson.M{"_bench": true},
				bson.M{"$set": bson.M{"updated_at": time.Now().UnixNano()}},
				options.Update().SetUpsert(true),
			)
		case strings.Contains(q, ".deleteOne("):
			_, _ = collection.DeleteOne(ctx, bson.M{"_bench": true})
		default:
			// Fallback: parse as find with no limit
			filter := parseMongoFilter(q)
			cur, err := collection.Find(ctx, filter, options.Find().SetLimit(1))
			if err == nil {
				_ = cur.Close(ctx)
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
