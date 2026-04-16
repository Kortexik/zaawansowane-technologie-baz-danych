package main

import (
	"benchmark/benchrunner"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	// Parse command line flags
	numTrials := flag.Int("trials", 3, "Number of trials for each benchmark")
	batchSize := flag.Int("batch", 10000, "Number of queries to run per benchmark")
	database := flag.String("db", "all", "Database to benchmark (mysql, postgres, mongodb, redis, all)")
	flag.Parse()

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║        Database Performance Benchmark Suite               ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Trials per test: %d\n", *numTrials)
	fmt.Printf("  Batch size: %d\n", *batchSize)
	fmt.Printf("  Target database: %s\n\n", *database)

	// Load database configurations
	dbConfig := benchrunner.DefaultDBConfig()
	
	// Create result set
	resultSet := &benchrunner.ResultSet{
		StartTime: time.Now(),
		Results:   []benchrunner.Result{},
	}

	// Run benchmarks based on selected database
	if *database == "all" || *database == "mysql" {
		if err := runMySQLBenchmarks(dbConfig.MySQL, *numTrials, *batchSize, resultSet); err != nil {
			log.Printf("MySQL benchmarks failed: %v", err)
		}
	}

	if *database == "all" || *database == "postgres" {
		if err := runPostgresBenchmarks(dbConfig.Postgres, *numTrials, *batchSize, resultSet); err != nil {
			log.Printf("PostgreSQL benchmarks failed: %v", err)
		}
	}

	if *database == "all" || *database == "mongodb" {
		if err := runMongoDBBenchmarks(dbConfig.MongoDB, *numTrials, *batchSize, resultSet); err != nil {
			log.Printf("MongoDB benchmarks failed: %v", err)
		}
	}

	if *database == "all" || *database == "redis" {
		if err := runRedisBenchmarks(dbConfig.Redis, *numTrials, *batchSize, resultSet); err != nil {
			log.Printf("Redis benchmarks failed: %v", err)
		}
	}

	// Finalize results
	resultSet.EndTime = time.Now()
	resultSet.TotalDuration = resultSet.EndTime.Sub(resultSet.StartTime)

	// Save results
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Saving results...")
	fmt.Println(strings.Repeat("=", 60))

	if err := os.MkdirAll("results", 0755); err != nil {
		log.Fatalf("Failed to create results directory: %v", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	jsonFile := fmt.Sprintf("results/benchmark_results_%s.json", timestamp)
	csvFile := fmt.Sprintf("results/benchmark_results_%s.csv", timestamp)

	if err := resultSet.SaveJSON(jsonFile); err != nil {
		log.Printf("Failed to save JSON results: %v", err)
	} else {
		fmt.Printf("✓ Results saved to %s\n", jsonFile)
	}

	if err := resultSet.SaveCSV(csvFile); err != nil {
		log.Printf("Failed to save CSV results: %v", err)
	} else {
		fmt.Printf("✓ Results saved to %s\n", csvFile)
	}

	fmt.Printf("\n✓ Benchmark completed in %s\n", resultSet.TotalDuration)
	fmt.Printf("✓ Total tests run: %d\n", len(resultSet.Results))
}

func runMySQLBenchmarks(config benchrunner.MySQLConfig, numTrials, batchSize int, resultSet *benchrunner.ResultSet) error {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Starting MySQL Benchmarks (with Index Comparison)")
	fmt.Println(strings.Repeat("=", 60))

	runner := benchrunner.NewMySQLRunner(config)
	if err := runner.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer runner.Close()

	// Data scales for Grade 4.0 requirement
	dataScales := []int{500000, 1000000, 10000000}
	
	// Define test scenarios
	scenarios := []struct {
		entity    string
		operation string
		queryFile string
	}{
		// Focus on SELECT operations for index testing
		{"users", "select", "queries/users_sql_select.sql"},
		{"categories", "select", "queries/categories_sql_select.sql"},
		{"products", "select", "queries/products_sql_select.sql"},
		{"addresses", "select", "queries/addresses_sql_select.sql"},
		{"orders", "select", "queries/orders_sql_select.sql"},
	}

	// Test without indexes first
	fmt.Println("\n📊 Testing WITHOUT indexes...")
	for _, dataScale := range dataScales {
		fmt.Printf("\n  Data Scale: %d records\n", dataScale)
		for _, scenario := range scenarios {
			results, err := runner.BenchmarkScenario(
				scenario.entity,
				scenario.operation,
				scenario.queryFile,
				batchSize,
				numTrials,
			)
			if err != nil {
				log.Printf("Scenario %s-%s failed: %v", scenario.entity, scenario.operation, err)
				continue
			}

			for _, r := range results {
				r.DataScale = dataScale
				r.WithIndex = false
				resultSet.Results = append(resultSet.Results, *r)
			}
		}
	}

	// Create indexes
	fmt.Println("\n📊 Creating indexes...")
	if err := runner.CreateIndexes(); err != nil {
		log.Printf("Index creation failed: %v", err)
	}

	// Test with indexes
	fmt.Println("\n📊 Testing WITH indexes...")
	for _, dataScale := range dataScales {
		fmt.Printf("\n  Data Scale: %d records\n", dataScale)
		for _, scenario := range scenarios {
			results, err := runner.BenchmarkScenario(
				scenario.entity,
				scenario.operation,
				scenario.queryFile,
				batchSize,
				numTrials,
			)
			if err != nil {
				log.Printf("Scenario %s-%s failed: %v", scenario.entity, scenario.operation, err)
				continue
			}

			for _, r := range results {
				r.DataScale = dataScale
				r.WithIndex = true
				resultSet.Results = append(resultSet.Results, *r)
			}
		}
	}

	// Drop indexes
	fmt.Println("\n📊 Cleaning up indexes...")
	if err := runner.DropIndexes(); err != nil {
		log.Printf("Index drop failed: %v", err)
	}

	return nil
}

func runPostgresBenchmarks(config benchrunner.PostgresConfig, numTrials, batchSize int, resultSet *benchrunner.ResultSet) error {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Starting PostgreSQL Benchmarks (with Index Comparison)")
	fmt.Println(strings.Repeat("=", 60))

	runner := benchrunner.NewPostgresRunner(config)
	if err := runner.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer runner.Close()

	// Data scales for Grade 4.0 requirement
	dataScales := []int{500000, 1000000, 10000000}
	
	// Define test scenarios - focus on SELECT for index testing
	scenarios := []struct {
		entity    string
		operation string
		queryFile string
	}{
		// Focus on SELECT operations for index testing
		{"users", "select", "queries/users_sql_select.sql"},
		{"categories", "select", "queries/categories_sql_select.sql"},
		{"products", "select", "queries/products_sql_select.sql"},
		{"addresses", "select", "queries/addresses_sql_select.sql"},
		{"orders", "select", "queries/orders_sql_select.sql"},
	}

	// Test without indexes first
	fmt.Println("\n📊 Testing WITHOUT indexes...")
	for _, dataScale := range dataScales {
		fmt.Printf("\n  Data Scale: %d records\n", dataScale)
		for _, scenario := range scenarios {
			results, err := runner.BenchmarkScenario(
				scenario.entity,
				scenario.operation,
				scenario.queryFile,
				batchSize,
				numTrials,
			)
			if err != nil {
				log.Printf("Scenario %s-%s failed: %v", scenario.entity, scenario.operation, err)
				continue
			}

			for _, r := range results {
				r.DataScale = dataScale
				r.WithIndex = false
				resultSet.Results = append(resultSet.Results, *r)
			}
		}
	}

	// Create indexes
	fmt.Println("\n📊 Creating indexes...")
	if err := runner.CreateIndexes(); err != nil {
		log.Printf("Index creation failed: %v", err)
	}

	// Test with indexes
	fmt.Println("\n📊 Testing WITH indexes...")
	for _, dataScale := range dataScales {
		fmt.Printf("\n  Data Scale: %d records\n", dataScale)
		for _, scenario := range scenarios {
			results, err := runner.BenchmarkScenario(
				scenario.entity,
				scenario.operation,
				scenario.queryFile,
				batchSize,
				numTrials,
			)
			if err != nil {
				log.Printf("Scenario %s-%s failed: %v", scenario.entity, scenario.operation, err)
				continue
			}

			for _, r := range results {
				r.DataScale = dataScale
				r.WithIndex = true
				resultSet.Results = append(resultSet.Results, *r)
			}
		}
	}

	// Drop indexes
	fmt.Println("\n📊 Cleaning up indexes...")
	if err := runner.DropIndexes(); err != nil {
		log.Printf("Index drop failed: %v", err)
	}

	return nil
}

func runMongoDBBenchmarks(config benchrunner.MongoDBConfig, numTrials, batchSize int, resultSet *benchrunner.ResultSet) error {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Starting MongoDB Benchmarks")
	fmt.Println(strings.Repeat("=", 60))

	runner := benchrunner.NewMongoDBRunner(config)
	if err := runner.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer runner.Close()

	// Data scales for Grade 4.0 requirement
	dataScales := []int{500000, 1000000, 10000000}
	
	// Define test scenarios
	scenarios := []struct {
		entity    string
		operation string
		queryFile string
	}{
		// Focus on core CRUD operations
		{"users", "find", "queries/users_mongo_find.js"},
		{"categories", "find", "queries/categories_mongo_find.js"},
		{"products", "find", "queries/products_mongo_find.js"},
	}

	for _, dataScale := range dataScales {
		fmt.Printf("\n  Data Scale: %d records\n", dataScale)
		for _, scenario := range scenarios {
			results, err := runner.BenchmarkScenario(
				scenario.entity,
				scenario.operation,
				scenario.queryFile,
				batchSize,
				numTrials,
			)
			if err != nil {
				log.Printf("Scenario %s-%s failed: %v", scenario.entity, scenario.operation, err)
				continue
			}

			for _, r := range results {
				r.DataScale = dataScale
				resultSet.Results = append(resultSet.Results, *r)
			}
		}
	}

	return nil
}

func runRedisBenchmarks(config benchrunner.RedisConfig, numTrials, batchSize int, resultSet *benchrunner.ResultSet) error {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Starting Redis Benchmarks")
	fmt.Println(strings.Repeat("=", 60))

	runner := benchrunner.NewRedisRunner(config)
	if err := runner.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer runner.Close()

	// Data scales for Grade 4.0 requirement
	dataScales := []int{500000, 1000000, 10000000}
	
	// Define test scenarios  
	scenarios := []struct {
		entity    string
		operation string
		queryFile string
	}{
		// Focus on core operations
		{"users", "get", "queries/users_redis_get.txt"},
		{"categories", "get", "queries/categories_redis_get.txt"},
		{"products", "get", "queries/products_redis_get.txt"},
		{"orders", "get", "queries/orders_redis_get.txt"},
	}

	for _, dataScale := range dataScales {
		fmt.Printf("\n  Data Scale: %d records\n", dataScale)
		for _, scenario := range scenarios {
			results, err := runner.BenchmarkScenario(
				scenario.entity,
				scenario.operation,
				scenario.queryFile,
				batchSize,
				numTrials,
			)
			if err != nil {
				log.Printf("Scenario %s-%s failed: %v", scenario.entity, scenario.operation, err)
				continue
			}

			for _, r := range results {
				r.DataScale = dataScale
				resultSet.Results = append(resultSet.Results, *r)
			}
		}
	}

	return nil
}

