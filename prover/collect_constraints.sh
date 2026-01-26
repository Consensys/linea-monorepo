#!/bin/bash

# Script to collect constraint counts for different round values
RESULTS_FILE="constraint_results.txt"
LOG_FILE="constraint_collection.log"

# Clear previous results
> "$RESULTS_FILE"
> "$LOG_FILE"

echo "Starting constraint collection for rounds 4-17..." | tee -a "$LOG_FILE"
echo "Results will be saved to: $RESULTS_FILE" | tee -a "$LOG_FILE"
echo "" | tee -a "$LOG_FILE"

# Array to store results
declare -A results

for round in {4..17}; do
    echo "========================================" | tee -a "$LOG_FILE"
    echo "Testing Round: $round" | tee -a "$LOG_FILE"
    echo "========================================" | tee -a "$LOG_FILE"
    
    # Update the Go file with current round value
    sed -i "s/if round == [0-9]\+/if round == $round/" protocol/wizard/gnark_verifier.go
    
    # Verify the change was made
    if grep -q "if round == $round" protocol/wizard/gnark_verifier.go; then
        echo "✓ Successfully updated code to round=$round" | tee -a "$LOG_FILE"
    else
        echo "✗ Failed to update code to round=$round" | tee -a "$LOG_FILE"
        continue
    fi
    
    # Run the benchmark and capture output
    echo "Running benchmark..." | tee -a "$LOG_FILE"
    timeout 300 go test -bench=BenchmarkCompilerWithSelfRecursionAndGnarkVerifier -run=^$ -v ./circuits 2>&1 | tee -a "$LOG_FILE" > temp_bench_output.txt
    
    # Extract constraint count
    constraint_count=$(grep "ccs number of constraints" temp_bench_output.txt | tail -1 | grep -oP '\d+' | tail -1)
    
    if [ -n "$constraint_count" ]; then
        echo "✓ Round $round: $constraint_count constraints" | tee -a "$LOG_FILE"
        results[$round]=$constraint_count
        echo "Round $round: $constraint_count" >> "$RESULTS_FILE"
    else
        echo "✗ Failed to extract constraint count for round $round" | tee -a "$LOG_FILE"
        echo "Round $round: FAILED" >> "$RESULTS_FILE"
    fi
    
    echo "" | tee -a "$LOG_FILE"
done

# Print summary
echo "========================================" | tee -a "$LOG_FILE"
echo "SUMMARY OF RESULTS" | tee -a "$LOG_FILE"
echo "========================================" | tee -a "$LOG_FILE"

for round in {4..17}; do
    if [ -n "${results[$round]}" ]; then
        printf "Round %2d: %'d constraints\n" $round ${results[$round]} | tee -a "$LOG_FILE"
    else
        printf "Round %2d: FAILED\n" $round | tee -a "$LOG_FILE"
    fi
done

echo "" | tee -a "$LOG_FILE"
echo "Data collection complete!" | tee -a "$LOG_FILE"
echo "Full results saved to: $RESULTS_FILE" | tee -a "$LOG_FILE"
echo "Full log saved to: $LOG_FILE" | tee -a "$LOG_FILE"

# Restore original value (optional - you can choose which round to leave it at)
# sed -i "s/if round == [0-9]\+/if round == 7/" protocol/wizard/gnark_verifier.go

