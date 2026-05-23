#!/bin/bash
# End-to-end benchmark experiment:
#   1. Wipe and restart Docker containers (MySQL/PostgreSQL/MongoDB/Redis).
#   2. Wait for them to be reachable.
#   3. Regenerate query files at the configured scale (numUsers in src/main.go).
#   4. Build the benchmark runner.
#   5. Run the benchmark — runner internally resets + loads data per scale
#      [500k, 1M, 10M] and tags every result row with `data_scale`.
#   6. Execute the analysis notebook in-place to refresh all charts.
#   7. Print where the results landed.
#
# All stdout/stderr is tee'd to logs/run_<timestamp>.log so you can leave it
# overnight and inspect what happened the next morning.
#
# Knobs (env or flags):
#   --trials N         number of trials per scenario (default 3)
#   --batch N          queries per benchmark for Phase B / fast paths (default 10000)
#   --batch-noidx N    smaller batch for Phase A / no-index scans (default 1000)
#   --db NAME          mysql|postgres|mongodb|redis|all (default all)
#   --skip-gen         do not regenerate query files (use cached ones)
#   --skip-notebook    do not execute the Jupyter notebook
#   --skip-reset       do not docker compose down -v (keep existing data on disk)
#
# Phase A (no indexes) runs only at the smallest scale; Phase B (with
# indexes) runs at every scale. This is hardcoded in main.go.

set -euo pipefail

TRIALS=${TRIALS:-3}
BATCH=${BATCH:-10000}
BATCH_NOIDX=${BATCH_NOIDX:-1000}
DATABASE=${DATABASE:-all}
SKIP_GEN=${SKIP_GEN:-0}
SKIP_NOTEBOOK=${SKIP_NOTEBOOK:-0}
SKIP_RESET=${SKIP_RESET:-0}

while [[ $# -gt 0 ]]; do
    case $1 in
        --trials) TRIALS="$2"; shift 2 ;;
        --batch) BATCH="$2"; shift 2 ;;
        --batch-noidx) BATCH_NOIDX="$2"; shift 2 ;;
        --db) DATABASE="$2"; shift 2 ;;
        --skip-gen) SKIP_GEN=1; shift ;;
        --skip-notebook) SKIP_NOTEBOOK=1; shift ;;
        --skip-reset) SKIP_RESET=1; shift ;;
        -h|--help)
            sed -n '1,30p' "$0"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Run '$0 --help' for usage."
            exit 1
            ;;
    esac
done

cd "$(dirname "$0")"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m'

mkdir -p logs
TS=$(date +%Y%m%d_%H%M%S)
LOG=logs/run_${TS}.log

# Mirror everything to the log file.
exec > >(tee -a "$LOG") 2>&1

START=$(date +%s)

cat <<BANNER
╔═══════════════════════════════════════════════════════════════╗
║          Full Benchmark Experiment — ${TS}             ║
╚═══════════════════════════════════════════════════════════════╝
  Trials:                  ${TRIALS}
  Batch (Phase B/fast):    ${BATCH}
  Batch (Phase A/no-idx):  ${BATCH_NOIDX}
  Database target:         ${DATABASE}
  Skip generation:         ${SKIP_GEN}
  Skip reset:              ${SKIP_RESET}
  Skip notebook:           ${SKIP_NOTEBOOK}
  Log file:                ${LOG}

Data scales: 500_000, 1_000_000, 10_000_000  (per src/cmd/run_benchmarks/main.go)
Phase A (no indexes) runs only at the smallest scale. Phase B (with
indexes) runs at every scale. With tuned MySQL/Postgres buffers, expected
runtime is ~3–5 hours.

BANNER

step() { echo -e "\n${BOLD}${YELLOW}▶ $1${NC}\n"; }
ok()   { echo -e "${GREEN}✓ $1${NC}"; }
fail() { echo -e "${RED}✗ $1${NC}"; }

# ── 1. Reset containers ──────────────────────────────────────────────────────
if [[ "$SKIP_RESET" == "0" ]]; then
    step "Step 1/6 — Reset Docker containers (volumes wiped)"
    docker compose down -v
    docker compose up -d
else
    step "Step 1/6 — Skipped reset; ensuring containers are up"
    docker compose up -d
fi

# Wait for each DB to actually accept connections.
step "Step 2/6 — Waiting for databases to become ready"
ready=0
for i in $(seq 1 60); do
    mysql_ok=0
    pg_ok=0
    mongo_ok=0
    redis_ok=0
    docker exec bench-mysql mysqladmin -ubench -pbench ping 2>/dev/null | grep -q "alive" && mysql_ok=1 || true
    docker exec bench-postgres pg_isready -U bench 2>/dev/null | grep -q "accepting" && pg_ok=1 || true
    docker exec bench-mongo mongosh --quiet --eval 'db.runCommand({ping:1}).ok' 2>/dev/null | grep -q "1" && mongo_ok=1 || true
    docker exec bench-redis redis-cli ping 2>/dev/null | grep -q "PONG" && redis_ok=1 || true
    if [[ "$mysql_ok" == "1" && "$pg_ok" == "1" && "$mongo_ok" == "1" && "$redis_ok" == "1" ]]; then
        ready=1
        break
    fi
    echo "  waiting ($i/60): mysql=$mysql_ok pg=$pg_ok mongo=$mongo_ok redis=$redis_ok"
    sleep 2
done
if [[ "$ready" != "1" ]]; then
    fail "Some databases never became ready. Check 'docker compose logs'."
    exit 1
fi
ok "All four databases reachable."

# ── 2. Regenerate query files ────────────────────────────────────────────────
if [[ "$SKIP_GEN" == "0" ]]; then
    step "Step 3/6 — Regenerate query files (this is slow at 10M scale)"
    rm -f queries/*.sql queries/*.js queries/*.txt
    (cd src && go run main.go)
    ok "Query files regenerated."
    du -sh queries/ || true
else
    step "Step 3/6 — Skipped (--skip-gen)"
fi

# ── 3. Build runner ──────────────────────────────────────────────────────────
step "Step 4/6 — Build benchmark runner"
mkdir -p bin
(cd src/cmd/run_benchmarks && go build -o ../../../bin/bench)
ok "bin/bench built."

# ── 4. Run benchmark ─────────────────────────────────────────────────────────
step "Step 5/6 — Run benchmark (this is the long phase)"
./bin/bench \
    -trials="$TRIALS" \
    -batch="$BATCH" \
    -batch-noidx="$BATCH_NOIDX" \
    -db="$DATABASE"
LATEST_CSV=$(ls -t results/benchmark_results_*.csv 2>/dev/null | head -1 || true)
LATEST_JSON=$(ls -t results/benchmark_results_*.json 2>/dev/null | head -1 || true)
ok "Benchmark complete: $LATEST_CSV"

# ── 5. Execute analysis notebook ─────────────────────────────────────────────
if [[ "$SKIP_NOTEBOOK" == "0" ]]; then
    step "Step 6/6 — Refresh analysis notebook"
    if [[ -d analysis/venv ]]; then
        # shellcheck disable=SC1091
        source analysis/venv/bin/activate
        if ! command -v jupyter >/dev/null 2>&1; then
            fail "jupyter not found in analysis/venv — installing required packages."
            pip install --quiet pandas matplotlib seaborn jupyter nbconvert
        fi
        (cd analysis && jupyter nbconvert \
            --to notebook \
            --execute benchmark_analysis.ipynb \
            --inplace \
            --ExecutePreprocessor.timeout=600)
        ok "Notebook refreshed → analysis/benchmark_analysis.ipynb"
        ok "Charts regenerated → analysis/chart_*.png"
    else
        fail "analysis/venv not found — skipping notebook step."
        echo "  To enable: cd analysis && python3 -m venv venv && source venv/bin/activate"
        echo "             && pip install pandas matplotlib seaborn jupyter nbconvert"
    fi
else
    step "Step 6/6 — Skipped (--skip-notebook)"
fi

END=$(date +%s)
ELAPSED=$((END - START))
HH=$((ELAPSED / 3600))
MM=$(( (ELAPSED % 3600) / 60 ))
SS=$((ELAPSED % 60))

cat <<DONE

╔═══════════════════════════════════════════════════════════════╗
║                       Experiment Complete                     ║
╚═══════════════════════════════════════════════════════════════╝
Total runtime: ${HH}h ${MM}m ${SS}s
Results CSV:   ${LATEST_CSV}
Results JSON:  ${LATEST_JSON}
Charts:        analysis/chart_*.png
Notebook:      analysis/benchmark_analysis.ipynb
Log:           ${LOG}
DONE
