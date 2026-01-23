#!/bin/bash

# Helper script for managing test cache and running conglomeration tests
# Usage: ./test-cached.sh [command]

CACHE_DIR="${LINEA_TEST_CACHE_DIR:-./debug/test-cache}"
TEST_TIMEOUT="${TEST_TIMEOUT:-30m}"

function show_help() {
    cat << EOF
Linea Conglomeration Test Cache Manager

Usage: $0 [command]

Commands:
    run                 Run the cached test (default)
    run-original        Run the original uncached test
    clear               Clear the cache
    clear-run           Clear cache and run test
    info                Show cache information
    help                Show this help message

Environment Variables:
    LINEA_TEST_CACHE_DIR    Cache directory (default: ./debug/test-cache)
    TEST_TIMEOUT            Test timeout (default: 30m)
    LINEA_CLEAR_CACHE       Set to 1 to clear cache before running

Examples:
    $0 run                      # Run with cache
    $0 clear-run                # Clear and run
    LINEA_CLEAR_CACHE=1 $0 run  # Same as clear-run
    $0 info                     # Show cache stats

EOF
}

function run_cached_test() {
    echo "Running cached conglomeration test..."
    echo "Cache directory: $CACHE_DIR"
    echo "Timeout: $TEST_TIMEOUT"
    echo ""
    
    go test -v -timeout "$TEST_TIMEOUT" -run TestConglomerationBasicCached 2>&1 | tee ./debug/debug_cached.log
    
    local exit_code=${PIPESTATUS[0]}
    if [ $exit_code -eq 0 ]; then
        echo ""
        echo "✓ Test passed!"
    else
        echo ""
        echo "✗ Test failed with exit code: $exit_code"
    fi
    return $exit_code
}

function run_original_test() {
    echo "Running original conglomeration test (no caching)..."
    echo "Timeout: $TEST_TIMEOUT"
    echo ""
    
    go test -v -timeout "$TEST_TIMEOUT" -run TestConglomerationBasic 2>&1 | tee ./debug/debug_original.log
    
    local exit_code=${PIPESTATUS[0]}
    if [ $exit_code -eq 0 ]; then
        echo ""
        echo "✓ Test passed!"
    else
        echo ""
        echo "✗ Test failed with exit code: $exit_code"
    fi
    return $exit_code
}

function clear_cache() {
    if [ -d "$CACHE_DIR" ]; then
        echo "Clearing cache directory: $CACHE_DIR"
        rm -rf "$CACHE_DIR"
        echo "✓ Cache cleared"
    else
        echo "Cache directory doesn't exist: $CACHE_DIR"
    fi
}

function show_cache_info() {
    echo "Cache Information"
    echo "================="
    echo "Cache directory: $CACHE_DIR"
    echo ""
    
    if [ -d "$CACHE_DIR" ]; then
        local cache_size=$(du -sh "$CACHE_DIR" 2>/dev/null | cut -f1)
        local file_count=$(find "$CACHE_DIR" -type f 2>/dev/null | wc -l)
        
        echo "Status: EXISTS"
        echo "Size: $cache_size"
        echo "Files: $file_count"
        echo ""
        echo "Cache files:"
        find "$CACHE_DIR" -type f -exec ls -lh {} \; 2>/dev/null | awk '{print "  " $9 " (" $5 ")"}'
    else
        echo "Status: DOES NOT EXIST"
        echo ""
        echo "Cache will be created on first test run."
    fi
}

# Main command dispatcher
case "${1:-run}" in
    run)
        run_cached_test
        ;;
    run-original)
        run_original_test
        ;;
    clear)
        clear_cache
        ;;
    clear-run)
        clear_cache
        echo ""
        run_cached_test
        ;;
    info)
        show_cache_info
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo "Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
