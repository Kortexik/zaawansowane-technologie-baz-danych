package benchrunner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRunner runs benchmarks for Redis
type RedisRunner struct {
	client *redis.Client
	config RedisConfig
}

// NewRedisRunner creates a new Redis benchmark runner
func NewRedisRunner(config RedisConfig) *RedisRunner {
	return &RedisRunner{
		config: config,
	}
}

// Connect establishes connection to Redis
func (r *RedisRunner) Connect() error {
	r.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", r.config.Host, r.config.Port),
		Password: r.config.Password,
		DB:       r.config.DB,
	})

	// Test connection
	ctx := context.Background()
	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("✓ Connected to Redis")
	return nil
}

// Close closes the database connection
func (r *RedisRunner) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// FlushAll wipes the entire Redis dataset so a fresh load can begin.
func (r *RedisRunner) FlushAll() error {
	ctx := context.Background()
	if err := r.client.FlushAll(ctx).Err(); err != nil {
		return fmt.Errorf("flushall: %w", err)
	}
	fmt.Println("✓ Redis dataset flushed")
	return nil
}

// LoadSetsFromFile streams the first `count` SET commands and executes them
// in 5k-command pipelines. Streaming keeps memory flat for 2 GB+ Redis files.
func (r *RedisRunner) LoadSetsFromFile(queryFile string, count int) error {
	ctx := context.Background()
	const batch = 5000
	pipe := r.client.Pipeline()
	inPipe := 0
	loaded := 0
	_, err := streamQueries(queryFile, count, func(line string) error {
		// Everything after "SET key " is the value (JSON may contain spaces).
		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 3 || strings.ToUpper(parts[0]) != "SET" {
			return nil
		}
		pipe.Set(ctx, parts[1], parts[2], 0)
		loaded++
		inPipe++
		if inPipe >= batch {
			if _, err := pipe.Exec(ctx); err != nil {
				fmt.Printf("  Load warning: %v\n", err)
			}
			pipe = r.client.Pipeline()
			inPipe = 0
		}
		return nil
	})
	if err != nil {
		return err
	}
	if _, err := pipe.Exec(ctx); err != nil {
		fmt.Printf("  Load warning: %v\n", err)
	}
	fmt.Printf("  ✓ Loaded %d keys from %s\n", loaded, queryFile)
	return nil
}

// RunQueryFile streams up to `limit` commands, then cycles them to fill
// exactly `limit` executions for timing.
func (r *RedisRunner) RunQueryFile(queryFile string, limit int) (*Result, error) {
	rawQueries := make([]string, 0, limit)
	if _, err := streamQueries(queryFile, limit, func(line string) error {
		rawQueries = append(rawQueries, line)
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

	// Execute queries and measure time
	ctx := context.Background()
	startTime := time.Now()

	for i, query := range queries {
		// Parse and execute Redis command
		parts := strings.Fields(query)
		if len(parts) == 0 {
			continue
		}

		cmd := strings.ToUpper(parts[0])
		args := parts[1:]

		switch cmd {
		case "SET":
			if len(args) >= 2 {
				if err := r.client.Set(ctx, args[0], args[1], 0).Err(); err != nil {
					if i < 5 {
						fmt.Printf("  Warning: Query %d failed: %v\n", i+1, err)
					}
				}
			}
		case "GET":
			if len(args) >= 1 {
				if _, err := r.client.Get(ctx, args[0]).Result(); err != nil && err.Error() != "redis: nil" {
					if i < 5 {
						fmt.Printf("  Warning: Query %d failed: %v\n", i+1, err)
					}
				}
			}
		case "DEL":
			if len(args) >= 1 {
				if _, err := r.client.Del(ctx, args[0]).Result(); err != nil {
					if i < 5 {
						fmt.Printf("  Warning: Query %d failed: %v\n", i+1, err)
					}
				}
			}
		case "INCR":
			if len(args) >= 1 {
				if _, err := r.client.Incr(ctx, args[0]).Result(); err != nil {
					if i < 5 {
						fmt.Printf("  Warning: Query %d failed: %v\n", i+1, err)
					}
				}
			}
		case "LPUSH":
			if len(args) >= 2 {
				values := make([]interface{}, len(args)-1)
				for i, v := range args[1:] {
					values[i] = v
				}
				if _, err := r.client.LPush(ctx, args[0], values...).Result(); err != nil {
					if i < 5 {
						fmt.Printf("  Warning: Query %d failed: %v\n", i+1, err)
					}
				}
			}
		case "LPOP":
			if len(args) >= 1 {
				if _, err := r.client.LPop(ctx, args[0]).Result(); err != nil && err.Error() != "redis: nil" {
					if i < 5 {
						fmt.Printf("  Warning: Query %d failed: %v\n", i+1, err)
					}
				}
			}
		case "HSET":
			if len(args) >= 3 {
				values := make(map[string]interface{})
				for j := 1; j < len(args); j += 2 {
					if j+1 < len(args) {
						values[args[j]] = args[j+1]
					}
				}
				if _, err := r.client.HSet(ctx, args[0], values).Result(); err != nil {
					if i < 5 {
						fmt.Printf("  Warning: Query %d failed: %v\n", i+1, err)
					}
				}
			}
		case "HGET":
			if len(args) >= 2 {
				if _, err := r.client.HGet(ctx, args[0], args[1]).Result(); err != nil && err.Error() != "redis: nil" {
					if i < 5 {
						fmt.Printf("  Warning: Query %d failed: %v\n", i+1, err)
					}
				}
			}
		case "SADD":
			if len(args) >= 2 {
				values := make([]interface{}, len(args)-1)
				for i, v := range args[1:] {
					values[i] = v
				}
				if _, err := r.client.SAdd(ctx, args[0], values...).Result(); err != nil {
					if i < 5 {
						fmt.Printf("  Warning: Query %d failed: %v\n", i+1, err)
					}
				}
			}
		case "SMEMBERS":
			if len(args) >= 1 {
				if _, err := r.client.SMembers(ctx, args[0]).Result(); err != nil {
					if i < 5 {
						fmt.Printf("  Warning: Query %d failed: %v\n", i+1, err)
					}
				}
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
func (r *RedisRunner) BenchmarkScenario(entity, operation, queryFile string, batchSize, numTrials int) ([]*Result, error) {
	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Printf("Redis Benchmark: %s - %s (batch: %d)\n", entity, strings.ToUpper(operation), batchSize)
	fmt.Printf("%s\n", strings.Repeat("=", 60))

	results := []*Result{}

	for trial := 1; trial <= numTrials; trial++ {
		fmt.Printf("\nTrial %d/%d:\n", trial, numTrials)

		result, err := r.RunQueryFile(queryFile, batchSize)
		if err != nil {
			return nil, fmt.Errorf("trial %d failed: %w", trial, err)
		}

		result.Database = "redis"
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
