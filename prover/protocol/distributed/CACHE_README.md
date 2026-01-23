# Test Caching for Conglomeration Tests

The conglomeration tests are very time-consuming (10+ minutes). To speed up iterative development and testing, a caching mechanism has been added.

## Quick Start

### Run the cached version of the test:
```bash
go test -timeout 30m -run TestConglomerationBasicCached
```

### Clear cache and run from scratch:
```bash
LINEA_CLEAR_CACHE=1 go test -timeout 30m -run TestConglomerationBasicCached
```

### Use a custom cache directory:
```bash
LINEA_TEST_CACHE_DIR=/path/to/cache go test -timeout 30m -run TestConglomerationBasicCached
```

## How It Works

The caching system saves intermediate compilation results to disk, allowing subsequent test runs to skip the expensive compilation phases:

1. **Wizard Compilation** - Initial protocol compilation
2. **Segment Compilation** - Module compilation (most expensive)
3. **Conglomeration Compilation** - Hierarchical conglomeration setup

### Cache Location

By default, cache files are stored in:
```
./debug/test-cache/
```

You can customize this with the `LINEA_TEST_CACHE_DIR` environment variable.

### Cache Invalidation

The cache is automatically invalidated when:
- Test parameters change (numRow, target weight, etc.)
- You manually clear it with `LINEA_CLEAR_CACHE=1`

To manually remove cache:
```bash
rm -rf ./debug/test-cache/
```

## What Gets Cached

Currently, the system attempts to cache:
- Compiled distributed wizards
- Segment compilation outputs
- Conglomeration compilation results

**Note:** Not all objects may be fully serializable. The caching system will gracefully fall back to recomputation if caching fails.

## Tests Available

### 1. TestConglomerationBasicCached
Cached version of the basic conglomeration test with small parameters.

### 2. TestConglomerationProverFileCached
Cached version for file-based testing with mainnet data.

### 3. Original Tests (unchanged)
- `TestConglomerationBasic` - Original test without caching
- `TestConglomerationProverFile` - Original file-based test
- `TestConglomerationProverDebug` - Debug test

## Limitations

Some structures may not be fully serializable with Go's `gob` encoder:
- Function pointers
- Channels
- Unexported struct fields from external packages
- Circuit constraints with complex closures

If you encounter serialization errors, the test will automatically fall back to full recompilation.

## Performance Impact

Expected speedup on subsequent runs:
- **First run:** Same as original (10-15 minutes)
- **Cached runs:** 30 seconds - 2 minutes (depending on what's cached)

The exact speedup depends on what part of the test you're iterating on.

## Debugging Cache Issues

Enable verbose logging:
```bash
go test -v -timeout 30m -run TestConglomerationBasicCached
```

Look for log messages:
- `[CACHE HIT]` - Successfully loaded from cache
- `[CACHE MISS]` - Computing from scratch
- `[CACHE WARNING]` - Non-fatal cache operation failure
- `[CACHE SAVED]` - Successfully saved to cache

## Advanced: Selective Caching

For more granular control, you can modify the cache keys in the test file to cache specific stages independently. This is useful when iterating on specific parts of the pipeline.

## Troubleshooting

### Cache corruption
If you get unexpected results, clear the cache:
```bash
LINEA_CLEAR_CACHE=1 go test -run TestConglomerationBasicCached
```

### Disk space
Cache files can be large (100MB+). Monitor your cache directory:
```bash
du -sh ./debug/test-cache/
```

### Permission errors
Ensure the cache directory is writable:
```bash
chmod -R u+w ./debug/test-cache/
```
