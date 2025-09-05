package fftdomain

import (
	"fmt"
	"runtime"
	"sync"
	"weak"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// domainCacheKey is the composite key for the cache.
// It uses a struct with comparable types as the map key.
type domainCacheKey struct {
	m              uint64
	gen            field.Element
	withPrecompute bool
}

var (
	domainCache = make(map[domainCacheKey]weak.Pointer[fft.Domain])
	domainMutex sync.Mutex
)

// NewDomainWithCache returns a subgroup with a power of 2 cardinality
// cardinality >= m
// shift: when specified, it's the element by which the set of root of unity is shifted.
// If domain is cached, it will directly returns the cached domain
func NewDomainWithCache(m uint64, withPrecompute bool, shift *field.Element) *fft.Domain {
	// Lock the mutex for the entire caching block.
	domainMutex.Lock()
	defer domainMutex.Unlock()

	// Compute the cache key.
	var gen field.Element
	if shift != nil {
		gen.Set(shift)
	} else {
		gen.SetUint64(field.MultiplicativeGen)
	}
	key := domainCacheKey{
		m:              m,
		gen:            gen,
		withPrecompute: withPrecompute,
	}

	// Return from cache if available.
	if domain, ok := domainCache[key]; ok {
		return domain.Value()
	}

	// Cache miss → create a new domain.
	var domain *fft.Domain
	switch {
	case shift != nil && withPrecompute:
		domain = fft.NewDomain(m, fft.WithShift(*shift))
	case withPrecompute:
		domain = fft.NewDomain(m)
	default:
		domain = fft.NewDomain(m, fft.WithoutPrecompute())
	}

	weakDomain := weak.Make(domain)

	// Store in cache.
	domainCache[key] = weakDomain

	// Add cleanup to remove from cache when domain is garbage collected.
	// IMPORTANT: Pass only the key data, not the domain itself
	runtime.AddCleanup(domain, func(key domainCacheKey) {
		fmt.Printf("AddCleanup called for data %v\n", key)

		domainMutex.Lock()
		delete(domainCache, key)
		domainMutex.Unlock()
	}, key)

	return domain
}
