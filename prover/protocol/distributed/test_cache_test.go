package distributed_test

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"

	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
)

// TestCache provides caching for expensive test operations
type TestCache struct {
	cacheDir string
}

// NewTestCache creates a new test cache in the specified directory
func NewTestCache(cacheDir string) *TestCache {
	if cacheDir == "" {
		cacheDir = filepath.Join(os.TempDir(), "linea-test-cache")
	}
	return &TestCache{cacheDir: cacheDir}
}

// ensureDir creates the cache directory if it doesn't exist
func (tc *TestCache) ensureDir() error {
	return os.MkdirAll(tc.cacheDir, 0o755)
}

// getCachePath returns the full path for a cache key
func (tc *TestCache) getCachePath(key string) string {
	return filepath.Join(tc.cacheDir, key+".cache")
}

// Save saves an object to the cache
func (tc *TestCache) Save(key string, obj interface{}) error {
	if err := tc.ensureDir(); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cachePath := tc.getCachePath(key)
	file, err := os.Create(cachePath)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(obj); err != nil {
		// If encoding fails, remove the partial file
		os.Remove(cachePath)
		return fmt.Errorf("failed to encode object: %w", err)
	}

	return nil
}

// Load loads an object from the cache
func (tc *TestCache) Load(key string, obj interface{}) error {
	cachePath := tc.getCachePath(key)

	file, err := os.Open(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("cache miss for key: %s", key)
		}
		return fmt.Errorf("failed to open cache file: %w", err)
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(obj); err != nil {
		return fmt.Errorf("failed to decode object: %w", err)
	}

	return nil
}

// Exists checks if a cache entry exists
func (tc *TestCache) Exists(key string) bool {
	cachePath := tc.getCachePath(key)
	_, err := os.Stat(cachePath)
	return err == nil
}

// Clear removes a specific cache entry
func (tc *TestCache) Clear(key string) error {
	cachePath := tc.getCachePath(key)
	if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear cache for key %s: %w", key, err)
	}
	return nil
}

// ClearAll removes all cache entries
func (tc *TestCache) ClearAll() error {
	return os.RemoveAll(tc.cacheDir)
}

// CacheableDistributedWizard wraps the DistributedWizard with only serializable fields
type CacheableDistributedWizard struct {
	// Store only the essential data that can be serialized
	// Note: This is a simplified version - you may need to adjust based on
	// what parts of DistributedWizard can actually be cached
	ModuleNames []string
	// Add other serializable fields as needed
}

// GetOrCompute retrieves from cache or computes and caches the result
func (tc *TestCache) GetOrCompute(key string, compute func() (interface{}, error)) (interface{}, error) {
	// Try to load from cache first
	var result interface{}
	if err := tc.Load(key, &result); err == nil {
		fmt.Printf("[CACHE HIT] Loaded %s from cache at %s\n", key, tc.getCachePath(key))
		return result, nil
	}

	fmt.Printf("[CACHE MISS] Computing %s...\n", key)
	// Cache miss, compute the result
	result, err := compute()
	if err != nil {
		return nil, err
	}

	// Save to cache for next time
	if err := tc.Save(key, result); err != nil {
		fmt.Printf("[CACHE WARNING] Failed to cache %s: %v\n", key, err)
		// Don't fail the test just because caching failed
	} else {
		fmt.Printf("[CACHE SAVED] Cached %s at %s\n", key, tc.getCachePath(key))
	}

	return result, nil
}

// SaveDistributedWizard saves a distributed wizard to cache
func (tc *TestCache) SaveDistributedWizard(key string, distWizard *distributed.DistributedWizard) error {
	return tc.Save(key, distWizard)
}

// LoadDistributedWizard loads a distributed wizard from cache
func (tc *TestCache) LoadDistributedWizard(key string) (*distributed.DistributedWizard, error) {
	var distWizard distributed.DistributedWizard
	if err := tc.Load(key, &distWizard); err != nil {
		return nil, err
	}
	return &distWizard, nil
}
