# Cache Implementation - In-Memory Approach

## What Changed

We've updated the caching approach to cache the `DistributedWizard` **right after `CompileSegments()`** and **before `Conglomerate()`**, as this is where the most expensive compilation happens (~6-7 minutes).

## Why In-Memory Caching?

The `DistributedWizard` structure contains:
- Function pointers
- Complex circuit constraints
- Non-serializable Go structures

These cannot be saved to disk using `gob` or other standard serialization methods. Therefore, we use an **in-memory cache** that persists within the same Go test process.

## How It Works

1. **Global Variables**: We use package-level variables to store the compiled wizard:
   ```go
   var (
       cachedDistWizardAfterCompile *distributed.DistributedWizard
       cacheCompileMutex            sync.Mutex
       cacheKey                     string
   )
   ```

2. **Cache Key**: Based on test parameters (numRow, targetWeight, etc.)

3. **Cache Check**: Before compiling, check if we already have a cached wizard with matching parameters

4. **Cache Hit**: If found, use the cached wizard immediately (saves ~6-7 minutes!)

5. **Cache Miss**: Compile from scratch and store in cache for next iteration

## Usage

### Running Tests Multiple Times in Same Session

**Important**: The cache only works within a **single test session**. Running `go test` twice will NOT share the cache because each invocation starts a fresh process.

To benefit from the cache, you need to run the test **multiple times in the same test session**:

```bash
# This script runs the test twice in the same session
./test-cached-twice.sh
```

Or manually in a Go test file, you could call the test function multiple times.

### Clearing the Cache

```bash
# Clear the in-memory cache
export LINEA_CLEAR_CACHE=1
go test -v -run TestConglomerationBasicCached
```

## Expected Performance

- **First Run** (cache miss): ~10-15 minutes (full compilation)
- **Second Run** (cache hit): ~4-8 minutes (skips CompileSegments, only runs Conglomerate + proofs)
- **Speedup**: Approximately 1.5-2x faster

## Limitations

1. **Process-Specific**: Cache doesn't persist across different `go test` invocations
2. **Memory Usage**: The cached wizard remains in memory
3. **Parameter Changes**: Changing test parameters invalidates the cache

## When to Use This

This caching approach is most useful when:
- Iterating on the latter parts of the test (proof generation, verification)
- Debugging conglomeration or proof aggregation logic
- Running the same test configuration multiple times during development

## Alternative Approaches (For Future)

If you need persistent cross-session caching, consider:
1. Serializing individual components separately (not the full wizard)
2. Using protocol-specific serialization methods if available
3. Caching at a different layer (e.g., compiled circuit artifacts)
