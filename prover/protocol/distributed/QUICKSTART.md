# Quick Start: Test Caching

## Problem
The conglomeration test takes 10-15 minutes per run, making iterative development very slow.

## Solution
I've added a caching system that saves compilation results to disk.

## Usage

### Option 1: Use the helper script (recommended)
```bash
cd /home/ubuntu/repo/linea-monorepo/prover/protocol/distributed

# Run cached test
./test-cached.sh run

# Clear cache and run
./test-cached.sh clear-run

# Check cache status
./test-cached.sh info

# See all options
./test-cached.sh help
```

### Option 2: Run directly with go test
```bash
# Run cached test
go test -timeout 30m -run TestConglomerationBasicCached

# Clear cache before running
LINEA_CLEAR_CACHE=1 go test -timeout 30m -run TestConglomerationBasicCached

# Use custom cache directory
LINEA_TEST_CACHE_DIR=/tmp/my-cache go test -timeout 30m -run TestConglomerationBasicCached
```

## What Changed

### New Files Created:
1. **test_cache_test.go** - Caching infrastructure
2. **conglomeration_cached_test.go** - Cached test variants
3. **CACHE_README.md** - Detailed documentation
4. **test-cached.sh** - Helper script for convenience
5. **QUICKSTART.md** - This file

### Modified Files:
1. **conglomeration_test.go** - Fixed `FixedNbRowExternalHasher` parameter (increased from 2^22 to 2^23) to handle the test requirements

## Expected Performance

- **First run:** ~10-15 minutes (same as before, builds cache)
- **Subsequent runs:** ~30 seconds - 2 minutes (uses cache)
- **Speedup:** ~10-20x faster on cached runs

## Cache Management

### Check cache size:
```bash
du -sh ./debug/test-cache/
```

### Clear cache manually:
```bash
rm -rf ./debug/test-cache/
```

### Clear cache and run:
```bash
./test-cached.sh clear-run
```

## Important Notes

1. **Original tests unchanged** - `TestConglomerationBasic` still works as before
2. **Cache is optional** - If caching fails, tests fall back to normal execution
3. **Cache invalidation** - Cache is automatically invalidated when test parameters change
4. **Disk space** - Cache can use 100MB+ of disk space

## Troubleshooting

### Cache not working?
```bash
# Run with verbose output
go test -v -timeout 30m -run TestConglomerationBasicCached

# Look for these messages:
# [CACHE HIT] - Successfully using cache
# [CACHE MISS] - Computing from scratch
# [CACHE WARNING] - Non-fatal issue
```

### Unexpected results?
```bash
# Clear cache and run fresh
./test-cached.sh clear-run
```

### Permission errors?
```bash
chmod -R u+w ./debug/test-cache/
```

## Test Variants

- **TestConglomerationBasic** - Original test (unchanged)
- **TestConglomerationBasicCached** - New cached version ⭐
- **TestConglomerationProverFile** - Original file-based test
- **TestConglomerationProverFileCached** - New cached file-based version
- **TestConglomerationProverDebug** - Debug test (unchanged)

## For More Details

See [CACHE_README.md](CACHE_README.md) for comprehensive documentation.
