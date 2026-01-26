#!/bin/bash

# Script to run benchmarks for different round values and collect constraint counts

RESULTS_FILE="round_benchmark_results.txt"
echo "Round Benchmark Results - $(date)" > $RESULTS_FILE
echo "========================================" >> $RESULTS_FILE
echo "" >> $RESULTS_FILE

for round in {4..17}; do
    echo "Testing round=$round..."
    echo "Round: $round" >> $RESULTS_FILE
    
    # Update the gnark_verifier.go file
    sed -i "s/if round == [0-9]\+/if round == $round/" /home/ubuntu/linea-monorepo/prover/protocol/wizard/gnark_verifier.go
    
    # Run the benchmark and capture output
    cd /home/ubuntu/linea-monorepo/prover
    OUTPUT=$(go test -benchmem -timeout=15m -run=^$ -bench ^BenchmarkCompilerWithSelfRecursionAndGnarkVerifier$ github.com/consensys/linea-monorepo/prover/protocol/compiler 2>&1)
    
    # Extract constraint count
    CONSTRAINTS=$(echo "$OUTPUT" | grep -oP 'ccs number of constraints: \K[0-9]+' | head -1)
    
    if [ -z "$CONSTRAINTS" ]; then
        echo "  Failed to extract constraints" | tee -a $RESULTS_FILE
        echo "$OUTPUT" >> $RESULTS_FILE
    else
        echo "  Constraints: $CONSTRAINTS" | tee -a $RESULTS_FILE
    fi
    
    echo "" >> $RESULTS_FILE
    echo "---" >> $RESULTS_FILE
    echo "" >> $RESULTS_FILE
done

echo "All benchmarks complete! Results saved to $RESULTS_FILE"
cat $RESULTS_FILE
