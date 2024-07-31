package mempool

import (
	"sync"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
)

// Pool pools the allocation for slices of [field.Element] of size `Size`. It
// should be used with great caution and every slice allocated via this pool
// must be manually freed and only once.
//
// Pool is used to reduce the number of allocation which can be significant
// when doing operations over field elements.
type Pool struct {
	Size int
	P    sync.Pool
}

// Create initializes the Pool with the given number of elements in it.
func Create(size int) *Pool {
	// Initializes the pool
	return &Pool{
		Size: size,
		P: sync.Pool{
			New: func() any {
				res := make([]field.Element, size)
				return &res
			},
		},
	}
}

// Prewarm the Pool by preallocating `nbPrewarm` in it.
func (p *Pool) Prewarm(nbPrewarm int) *Pool {
	prewarmed := make([]field.Element, p.Size*nbPrewarm)
	parallel.Execute(nbPrewarm, func(start, stop int) {
		for i := start; i < stop; i++ {
			vec := prewarmed[i*p.Size : (i+1)*p.Size]
			p.P.Put(&vec)
		}
	})
	return p
}

// Alloc returns a vector allocated from the pool. Vector allocated via the
// pool should ideally be returned to the pool. If not, they are still going to
// be picked up by the GC.
func (p *Pool) Alloc() *[]field.Element {
	res := p.P.Get().(*[]field.Element)
	return res
}

// Free returns an object to the pool. It must never be called twice over
// the same object or undefined behaviours are going to arise. It is fine to
// pass objects allocated to outside of the pool as long as they have the right
// dimension.
func (p *Pool) Free(vec *[]field.Element) error {
	// Check the vector has the right size
	if len(*vec) != p.Size {
		utils.Panic("expected size %v, expected %v", len(*vec), p.Size)
	}

	p.P.Put(vec)

	return nil
}

// ExtractCheckOptionalStrict returns
//   - p[0], true if the expectedSize matches the one of the provided pool
//   - nil, false if no `p` is provided
//   - panic if the assigned size of the pool does not match
//   - panic if the caller provides `nil` as argument for `p`
//
// This is used to unwrap a [Pool] that is commonly passed to functions as an
// optional variadic parameter and at the same time validating that the pool
// object has the right size.
func ExtractCheckOptionalStrict(expectedSize int, p ...*Pool) (pool *Pool, ok bool) {
	// Checks if there is a pool
	hasPool := len(p) > 0 && p[0] != nil
	if hasPool {
		pool = p[0]
	}

	// Sanity-check that the size of the pool is actually what we expected
	if hasPool && pool.Size != expectedSize {
		utils.Panic("pooled vector size are %v, but required %v", pool.Size, expectedSize)
	}

	return pool, hasPool
}

// ExtractCheckOptionalSoft returns
//   - p[0], true if the expectedSize matches the one of the provided pool
//   - nil, false if no `p` is provided
//   - nil, false if the length of the vector does not match the one of the pool
//   - panic if the caller provides `nil` as argument for `p`
//
// This is used to unwrap a [Pool] that is commonly passed to functions as an
// optional variadic parameter.
func ExtractCheckOptionalSoft(expectedSize int, p ...*Pool) (pool *Pool, ok bool) {
	// Checks if there is a pool
	hasPool := len(p) > 0
	if hasPool {
		pool = p[0]
	}

	// Sanity-check that the size of the pool is actually what we expected
	if hasPool && pool.Size != expectedSize {
		return nil, false
	}

	return pool, hasPool
}
