package benchrunner

import "time"

// Config holds benchmark configuration
type Config struct {
	NumTrials      int
	BatchSizes     []int
	DataScales     []int // Number of records: 500000, 1000000, 10000000
	Timeout        time.Duration
	TestIndexes    bool // Test before and after index creation
	AnalyzePlans   bool // Analyze query execution plans
}

// DefaultConfig returns default benchmark configuration
func DefaultConfig() *Config {
	return &Config{
		NumTrials:    3,
		BatchSizes:   []int{10000, 100000, 1000000},
		DataScales:   []int{500000, 1000000, 10000000},
		Timeout:      60 * time.Minute,
		TestIndexes:  true,
		AnalyzePlans: true,
	}
}

// DBConfig holds database connection configuration
type DBConfig struct {
	MySQL    MySQLConfig
	Postgres PostgresConfig
	MongoDB  MongoDBConfig
	Redis    RedisConfig
}

type MySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

type MongoDBConfig struct {
	Host       string
	Port       int
	User       string
	Password   string
	Database   string
	AuthSource string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// DefaultDBConfig returns default database configurations
func DefaultDBConfig() *DBConfig {
	return &DBConfig{
		MySQL: MySQLConfig{
			Host:     "localhost",
			Port:     3306,
			User:     "bench",
			Password: "bench",
			Database: "benchmark",
		},
		Postgres: PostgresConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "bench",
			Password: "bench",
			Database: "benchmark",
			SSLMode:  "disable",
		},
		MongoDB: MongoDBConfig{
			Host:       "localhost",
			Port:       27017,
			User:       "bench",
			Password:   "bench",
			Database:   "benchmark",
			AuthSource: "admin",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},
	}
}

