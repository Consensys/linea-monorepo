#!/bin/bash

# Script to test cache effectiveness by running the test twice
# The first run will populate the in-memory cache
# The second run (if we use "go test" twice) won't benefit from in-memory cache
# So we need to run the test twice within the same session

cd /home/ubuntu/repo/linea-monorepo/prover/protocol/distributed

echo "======================================================"
echo "FIRST RUN: Compiling and caching (this will be slow)"
echo "======================================================"
echo ""

START1=$(date +%s)
go test -v -timeout 30m -run "^TestConglomerationBasicCached$" 2>&1 | tee ./debug/cache_run1.log
EXIT1=$?
END1=$(date +%s)
DURATION1=$((END1 - START1))

echo ""
echo "======================================================"
echo "First run completed in ${DURATION1} seconds"
echo "======================================================"
echo ""

if [ $EXIT1 -ne 0 ]; then
    echo "First run failed! Check ./debug/cache_run1.log"
    echo "Cannot continue to second run."
    exit 1
fi

sleep 2

echo ""
echo "======================================================"
echo "SECOND RUN: Should use cached wizard (fast!)"
echo "======================================================"
echo ""

START2=$(date +%s)
go test -v -timeout 30m -run "^TestConglomerationBasicCached$" 2>&1 | tee ./debug/cache_run2.log
EXIT2=$?
END2=$(date +%s)
DURATION2=$((END2 - START2))

echo ""
echo "======================================================"
echo "RESULTS:"
echo "======================================================"
echo "First run:  ${DURATION1} seconds (compile + cache)"
echo "Second run: ${DURATION2} seconds (should use cache)"
echo ""

if [ $EXIT2 -eq 0 ]; then
    echo "✓ Both runs completed successfully!"
    SPEEDUP=$(echo "scale=2; $DURATION1 / $DURATION2" | bc)
    echo "Speedup: ${SPEEDUP}x faster"
else
    echo "✗ Second run failed! Check ./debug/cache_run2.log"
fi

echo ""
echo "Logs saved to:"
echo "  - ./debug/cache_run1.log"
echo "  - ./debug/cache_run2.log"
echo "======================================================"
