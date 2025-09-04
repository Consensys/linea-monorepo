package fftdomain

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/stretchr/testify/require"
)

func TestNewDomainGenerateCache(t *testing.T) {
	assert := require.New(t)

	// Reset cache to ensure tests are isolated.
	domainMutex.Lock()
	domainCache = make(map[domainCacheKey]*fft.Domain)
	domainMutex.Unlock()

	var gen field.Element
	gen.SetUint64(field.MultiplicativeGen)

	// Case 1: withPrecompute == true, no shift
	key1 := domainCacheKey{
		m:              256,
		gen:            gen,
		withPrecompute: true,
	}
	assert.Nil(domainCache[key1], "Before domain generation, domainCache[key1] should be nil")

	domain1 := NewDomainWithCache(256, true, nil)
	expected1 := fft.NewDomain(256)

	assert.Equal(domain1, expected1, "domain1 should equal expected1")
	assert.Equal(domain1, domainCache[key1], "domain1 should be stored in cache")

	// Case 2: withPrecompute == true, with shift
	shift := field.NewElement(5)
	key2 := domainCacheKey{
		m:              512,
		gen:            shift,
		withPrecompute: true,
	}
	assert.Nil(domainCache[key2], "Before domain generation, domainCache[key2] should be nil")

	domain2 := NewDomainWithCache(512, true, &shift)
	expected2 := fft.NewDomain(512, fft.WithShift(shift))

	assert.Equal(domain2, expected2, "domain2 should equal expected2")
	assert.Equal(domain2, domainCache[key2], "domain2 should be stored in cache")
	assert.NotSame(domain1, domain2, "Different domains should not be the same pointer")

	// Case 3: withPrecompute == false
	key3 := domainCacheKey{
		m:              256,
		gen:            gen,
		withPrecompute: false,
	}
	assert.Nil(domainCache[key3], "Before domain generation, domainCache[key3] should be nil")

	domain3 := NewDomainWithCache(256, false, nil)
	expected3 := fft.NewDomain(256, fft.WithoutPrecompute())

	assert.Equal(domain3, expected3, "domain3 should equal expected3")
	assert.Equal(domain3, domainCache[key3], "domain3 should be stored in cache")
}

func BenchmarkNewDomainCache(b *testing.B) {
	b.Run("CacheHit", func(b *testing.B) {
		// Ensure first call populates cache
		NewDomainWithCache(1<<20, true, nil)

		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			NewDomainWithCache(1<<20, true, nil)
		}
	})

	b.Run("CacheMiss", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			fft.NewDomain(1 << 20)
		}
	})
}
