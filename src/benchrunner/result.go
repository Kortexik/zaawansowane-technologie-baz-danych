package benchrunner

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Result represents a single benchmark trial result
type Result struct {
	Database       string        `json:"database"`
	Entity         string        `json:"entity"`
	Operation      string        `json:"operation"`
	BatchSize      int           `json:"batch_size"`
	DataScale      int           `json:"data_scale"`
	Trial          int           `json:"trial"`
	NumQueries     int           `json:"num_queries"`
	TotalTime      time.Duration `json:"total_time_ns"`
	TotalTimeMs    float64       `json:"total_time_ms"`
	AvgTimePerOp   time.Duration `json:"avg_time_per_op_ns"`
	AvgTimePerOpMs float64       `json:"avg_time_per_op_ms"`
	QueriesPerSec  float64       `json:"queries_per_second"`
	WithIndex      bool          `json:"with_index"`
	QueryPlan      string        `json:"query_plan,omitempty"`
	Timestamp      time.Time     `json:"timestamp"`
}

// AggregatedResult represents averaged results across multiple trials
type AggregatedResult struct {
	Database         string  `json:"database"`
	Entity           string  `json:"entity"`
	Operation        string  `json:"operation"`
	BatchSize        int     `json:"batch_size"`
	DataScale        int     `json:"data_scale"`
	NumTrials        int     `json:"num_trials"`
	AvgTotalTimeMs   float64 `json:"avg_total_time_ms"`
	AvgQPS           float64 `json:"avg_queries_per_second"`
	MinTimeMs        float64 `json:"min_time_ms"`
	MaxTimeMs        float64 `json:"max_time_ms"`
	StdDevMs         float64 `json:"std_dev_ms"`
	AvgTimePerOpMs   float64 `json:"avg_time_per_op_ms"`
	WithIndex        bool    `json:"with_index"`
}

// ResultSet holds all benchmark results
type ResultSet struct {
	Results            []Result           `json:"results"`
	AggregatedResults  []AggregatedResult `json:"aggregated_results"`
	StartTime          time.Time          `json:"start_time"`
	EndTime            time.Time          `json:"end_time"`
	TotalDuration      time.Duration      `json:"total_duration_ns"`
}

// NewResult creates a new benchmark result
func NewResult(database, entity, operation string, batchSize, dataScale, trial, numQueries int, totalTime time.Duration) *Result {
	totalTimeMs := float64(totalTime.Milliseconds())
	avgTimePerOp := time.Duration(0)
	avgTimePerOpMs := 0.0
	qps := 0.0

	if numQueries > 0 {
		avgTimePerOp = totalTime / time.Duration(numQueries)
		avgTimePerOpMs = totalTimeMs / float64(numQueries)
		qps = float64(numQueries) / totalTime.Seconds()
	}

	return &Result{
		Database:       database,
		Entity:         entity,
		Operation:      operation,
		BatchSize:      batchSize,
		DataScale:      dataScale,
		Trial:          trial,
		NumQueries:     numQueries,
		TotalTime:      totalTime,
		TotalTimeMs:    totalTimeMs,
		AvgTimePerOp:   avgTimePerOp,
		AvgTimePerOpMs: avgTimePerOpMs,
		QueriesPerSec:  qps,
		WithIndex:      false,
		Timestamp:      time.Now(),
	}
}

// SaveJSON saves results to a JSON file
func (rs *ResultSet) SaveJSON(filename string) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(rs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// SaveCSV saves results to a CSV file
func (rs *ResultSet) SaveCSV(filename string) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"database", "entity", "operation", "batch_size", "data_scale", "trial",
		"num_queries", "total_time_ms", "avg_time_per_op_ms", "queries_per_second", "with_index",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write data
	for _, r := range rs.Results {
		row := []string{
			r.Database,
			r.Entity,
			r.Operation,
			fmt.Sprintf("%d", r.BatchSize),
			fmt.Sprintf("%d", r.DataScale),
			fmt.Sprintf("%d", r.Trial),
			fmt.Sprintf("%d", r.NumQueries),
			fmt.Sprintf("%.2f", r.TotalTimeMs),
			fmt.Sprintf("%.4f", r.AvgTimePerOpMs),
			fmt.Sprintf("%.2f", r.QueriesPerSec),
			fmt.Sprintf("%v", r.WithIndex),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

