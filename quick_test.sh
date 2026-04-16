#!/bin/bash

# Szybki test benchmarku (mała ilość danych)

echo "🚀 Quick Benchmark Test"
echo "======================="
echo ""
echo "This will run a quick test with:"
echo "  - 1000 queries per test"
echo "  - 1 trial per test"
echo "  - Only MySQL and PostgreSQL"
echo ""

# Sprawdź czy bazy działają
if ! docker ps | grep -q "bench-mysql"; then
    echo "Starting databases..."
    docker-compose up -d
    sleep 5
fi

# Kompiluj jeśli potrzeba
if [ ! -f "bin/bench" ]; then
    echo "Building benchmark runner..."
    mkdir -p bin
    cd src/cmd/run_benchmarks
    go build -o ../../../bin/bench
    cd ../../..
fi

# Uruchom szybki test
echo ""
echo "Running quick test..."
echo ""

./bin/bench -trials=1 -batch=1000 -db=mysql

echo ""
echo "✓ Quick test complete!"
echo "Check results/ directory for output"

