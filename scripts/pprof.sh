#!/bin/bash
# FlyFleet pprof Helper Script
# Usage: ./scripts/pprof.sh [profile_type] [options]
# Examples:
#   ./scripts/pprof.sh cpu          # CPU profile (30s)
#   ./scripts/pprof.sh heap         # Memory heap
#   ./scripts/pprof.sh goroutine    # Goroutine dump
#   ./scripts/pprof.sh cpu -seconds 60  # Custom duration

set -e

PPROF_URL=${PPROF_URL:-"http://localhost:6060"}
OUTPUT_DIR=${OUTPUT_DIR:-"./pprof-results"}
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

profile_type=$1
shift

case "$profile_type" in
    cpu)
        echo "📊 Capturing CPU profile (30s)..."
        SECONDS=${1:-30}
        go tool pprof -http=:8081 "${PPROF_URL}/debug/pprof/profile?seconds=${SECONDS}" 2>/dev/null &
        sleep $SECONDS
        curl -s "${PPROF_URL}/debug/pprof/profile?seconds=${SECONDS}" -o "${OUTPUT_DIR}/cpu-${TIMESTAMP}.prof"
        echo "✅ CPU profile saved to: ${OUTPUT_DIR}/cpu-${TIMESTAMP}.prof"
        echo "   View with: go tool pprof -http=:8081 ${OUTPUT_DIR}/cpu-${TIMESTAMP}.prof"
        ;;

    heap)
        echo "🧠 Capturing memory heap profile..."
        curl -s "${PPROF_URL}/debug/pprof/heap" -o "${OUTPUT_DIR}/heap-${TIMESTAMP}.prof"
        echo "✅ Heap profile saved to: ${OUTPUT_DIR}/heap-${TIMESTAMP}.prof"
        echo "   View with: go tool pprof -http=:8081 ${OUTPUT_DIR}/heap-${TIMESTAMP}.prof"
        ;;

    goroutine|threads)
        echo "👥 Capturing goroutine dump..."
        curl -s "${PPROF_URL}/debug/pprof/goroutine?debug=1" -o "${OUTPUT_DIR}/goroutines-${TIMESTAMP}.txt"
        COUNT=$(grep -c "^goroutine " "${OUTPUT_DIR}/goroutines-${TIMESTAMP}.txt" 2>/dev/null || echo "0")
        echo "✅ Goroutine dump saved to: ${OUTPUT_DIR}/goroutines-${TIMESTAMP}.txt (${COUNT} goroutines)"
        ;;

    alloc)
        echo "📈 Capturing allocation profile..."
        curl -s "${PPROF_URL}/debug/pprof/allocs" -o "${OUTPUT_DIR}/allocs-${TIMESTAMP}.prof"
        echo "✅ Alloc profile saved to: ${OUTPUT_DIR}/allocs-${TIMESTAMP}.prof"
        ;;

    block)
        echo "🔒 Capturing block profile (mutex contention)..."
        curl -s "${PPROF_URL}/debug/pprof/block" -o "${OUTPUT_DIR}/block-${TIMESTAMP}.prof"
        echo "✅ Block profile saved to: ${OUTPUT_DIR}/block-${TIMESTAMP}.prof"
        ;;

    mutex)
        echo "🔒 Capturing mutex profile..."
        curl -s "${PPROF_URL}/debug/pprof/mutex" -o "${OUTPUT_DIR}/mutex-${TIMESTAMP}.prof"
        echo "✅ Mutex profile saved to: ${OUTPUT_DIR}/mutex-${TIMESTAMP}.prof"
        ;;

    baseline)
        echo "📋 Capturing baseline profiles (cpu, heap, goroutines)..."
        echo "   This may take ~30 seconds for CPU profile..."
        ./scripts/pprof.sh cpu 30
        ./scripts/pprof.sh heap
        ./scripts/pprof.sh goroutine
        echo "✅ Baseline profiles captured in ${OUTPUT_DIR}/"
        ;;

    compare)
        # Compare two profiles
        BASE=$2
        COMPARE=$3
        if [ -z "$BASE" ] || [ -z "$COMPARE" ]; then
            echo "Usage: $0 compare <base.prof> <compare.prof>"
            exit 1
        fi
        go tool pprof -http=:8081 "$BASE" "$COMPARE"
        ;;

    list)
        echo "📂 Available profiles in ${OUTPUT_DIR}:"
        ls -lah "$OUTPUT_DIR" 2>/dev/null || echo "  (no profiles found)"
        ;;

    *)
        echo "FlyFleet pprof Helper"
        echo ""
        echo "Usage: $0 <profile_type> [options]"
        echo ""
        echo "Profile Types:"
        echo "  cpu [seconds]    - CPU profile (default: 30s)"
        echo "  heap             - Memory heap snapshot"
        echo "  goroutine        - Goroutine stack dump"
        echo "  alloc            - Allocation profile"
        echo "  block            - Block (contention) profile"
        echo "  mutex            - Mutex contention profile"
        echo "  baseline         - Capture all baseline profiles"
        echo "  compare <base> <compare> - Compare two profiles in browser"
        echo "  list             - List saved profiles"
        echo ""
        echo "Environment Variables:"
        echo "  PPROF_URL=http://localhost:6060  - pprof endpoint"
        echo "  OUTPUT_DIR=./pprof-results      - Output directory"
        echo ""
        echo "Examples:"
        echo "  $0 cpu 60                        # 60s CPU profile"
        echo "  $0 heap                          # Heap snapshot"
        echo "  $0 baseline                      # Capture all baselines"
        echo "  $0 compare cpu-base.prof cpu-new.prof"
        exit 1
        ;;
esac
