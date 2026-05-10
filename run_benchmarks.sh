#!/bin/bash

# Skrypt do uruchamiania pełnego benchmarku baz danych

set -e

echo "╔════════════════════════════════════════════════════════════╗"
echo "║     Database Benchmark - Automated Test Suite             ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Kolory
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Funkcja do sprawdzania czy kontener działa
check_container() {
    local container=$1
    if docker ps | grep -q "$container"; then
        echo -e "${GREEN}✓${NC} $container is running"
        return 0
    else
        echo -e "${RED}✗${NC} $container is not running"
        return 1
    fi
}

# Krok 1: Sprawdź czy bazy danych działają
echo "Step 1: Checking database containers..."
echo "----------------------------------------"

all_running=true
check_container "bench-mysql" || all_running=false
check_container "bench-postgres" || all_running=false
check_container "bench-mongo" || all_running=false
check_container "bench-redis" || all_running=false

if [ "$all_running" = false ]; then
    echo ""
    echo -e "${YELLOW}Some containers are not running. Starting them...${NC}"
    docker-compose up -d
    echo "Waiting 10 seconds for databases to initialize..."
    sleep 10
fi

echo ""

# Krok 2: Sprawdź czy pliki z zapytaniami istnieją
echo "Step 2: Checking query files..."
echo "----------------------------------------"

if [ ! -f "queries/users_sql_insert.sql" ]; then
    echo -e "${YELLOW}Query files not found. Generating...${NC}"
    cd src
    if [ ! -f "benchmark" ]; then
        echo "Building data generator..."
        go build -o benchmark
    fi
    echo "Generating query files (this may take a few minutes)..."
    ./benchmark
    cd ..
    echo -e "${GREEN}✓${NC} Query files generated"
else
    echo -e "${GREEN}✓${NC} Query files exist"
fi

echo ""

# Krok 3: Kompiluj benchmark runner
echo "Step 3: Building benchmark runner..."
echo "----------------------------------------"

if [ ! -f "bin/bench" ]; then
    echo "Compiling benchmark runner..."
    cd src/cmd/run_benchmarks
    go build -o ../../../bin/bench
    cd ../../..
fi
echo -e "${GREEN}✓${NC} Benchmark runner ready"

echo ""

# Krok 4: Uruchom benchmarki
echo "Step 4: Running benchmarks..."
echo "----------------------------------------"

# Parsuj argumenty
TRIALS=3
BATCH=10000
DATABASE="all"

while [[ $# -gt 0 ]]; do
    case $1 in
        --trials)
            TRIALS="$2"
            shift 2
            ;;
        --batch)
            BATCH="$2"
            shift 2
            ;;
        --db)
            DATABASE="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--trials N] [--batch N] [--db mysql|postgres|all]"
            exit 1
            ;;
    esac
done

echo "Configuration:"
echo "  Trials: $TRIALS"
echo "  Batch size: $BATCH"
echo "  Database: $DATABASE"
echo ""

./bin/bench -trials=$TRIALS -batch=$BATCH -db=$DATABASE

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║                  Benchmark Complete!                       ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "Results saved in: results/"
echo ""
echo "To view results:"
echo "  - JSON: cat results/benchmark_results_*.json | jq"
echo "  - CSV:  cat results/benchmark_results_*.csv"
echo ""

