package mempoolext

import (
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/parallel"
	"sync"
)

// FromSyncPool pools the allocation for slices of [fext.Element] of size `Size`.
// It should be used with great caution and every slice allocated via this pool
// must be manually freed and only once.
//
// FromSyncPool is used to reduce the number of allocation which can be significant
// when doing operations over field elements.
type FromSyncPool struct {
	size int
	P    sync.Pool
}

// CreateFromSyncPool initializes the Pool with the given number of elements in it.
func CreateFromSyncPool(size int) *FromSyncPool {
	// Initializes the pool
	return &FromSyncPool{
		size: size,
		P: sync.Pool{
			New: func() any {
				res := make([]fext.Element, size)
				return &res
			},
		},
	}
}

// Prewarm the Pool by preallocating `nbPrewarm` in it.
func (p *FromSyncPool) Prewarm(nbPrewarm int) MemPool {
	prewarmed := make([]fext.Element, p.size*nbPrewarm)
	parallel.Execute(nbPrewarm, func(start, stop int) {
		for i := start; i < stop; i++ {
			vec := prewarmed[i*p.size : (i+1)*p.size]
			p.P.Put(&vec)
		}
	})
	return p
}

// Alloc returns a vector allocated from the pool. Vector allocated via the
// pool should ideally be returned to the pool. If not, they are still going to
// be picked up by the GC.
func (p *FromSyncPool) Alloc() *[]fext.Element {
	res := p.P.Get().(*[]fext.Element)
	return res
}

// Free returns an object to the pool. It must never be called twice over
// the same object or undefined behaviours are going to arise. It is fine to
// pass objects allocated to outside of the pool as long as they have the right
// dimension.
func (p *FromSyncPool) Free(vec *[]fext.Element) error {
	// Check the vector has the right size
	if len(*vec) != p.size {
		utils.Panic("expected size %v, expected %v", len(*vec), p.Size())
	}

	p.P.Put(vec)

	return nil
}

func (p *FromSyncPool) Size() int {
	return p.size
}
