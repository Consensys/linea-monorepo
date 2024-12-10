package smartvectors

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/maths/common/mempool"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// It's normal vector in a nutshell
type Regular[T fmt.Stringer] []T

// Instanstiate a new regular from a slice. Returns a pointer so that the result
// can be reused without referencing as a SmartVector.
func NewRegular[T fmt.Stringer](v []T) *Regular[T] {
	assertStrictPositiveLen(len(v))
	res := Regular[T](v)
	return &res
}

// Returns the length of the regular vector
func (r *Regular[T]) Len() int { return len(*r) }

// Returns a particular element of the vector
func (r *Regular[T]) Get(n int) T { return (*r)[n] }

// Returns a subvector of the regular
func (r *Regular[T]) SubVector(start, stop int) SmartVector[T] {
	if start > stop {
		utils.Panic("Negative length are not allowed")
	}
	if start == stop {
		utils.Panic("Subvector of zero lengths are not allowed")
	}
	res := Regular[T]((*r)[start:stop])
	return &res
}

// Rotates the vector into a new one
func (r *Regular[T]) RotateRight(offset int) SmartVector[T] {
	resSlice := make(Regular[T], r.Len())

	if offset == 0 {
		copy(resSlice, *r)
		return &resSlice
	}

	if offset > 0 {
		// v and w may be the same vector thus we should use a
		// separate leftover buffer for temporary memory buffers.
		cutAt := len(*r) - offset
		leftovers := vector.DeepCopy((*r)[cutAt:])
		copy(resSlice[offset:], (*r)[:cutAt])
		copy(resSlice[:offset], leftovers)
		return &resSlice
	}

	if offset < 0 {
		glueAt := len(*r) + offset
		leftovers := vector.DeepCopy((*r)[:-offset])
		copy(resSlice[:glueAt], (*r)[-offset:])
		copy(resSlice[glueAt:], leftovers)
		return &resSlice
	}

	panic("unreachable")
}

func (r *Regular[T]) WriteInSlice(s []T) {
	assertHasLength(len(s), len(*r))
	copy(s, *r)
}

func (r *Regular[T]) Pretty() string {
	return fmt.Sprintf("Regular[%v]", vector.Prettify(*r))
}

func processRegularOnly[T fmt.Stringer](op operator, svecs []SmartVector[T], coeffs []int, p ...mempool.MemPool[T]) (result *Pooled[T], numMatches int) {

	length := svecs[0].Len()

	pool, hasPool := mempool.ExtractCheckOptionalStrict(length, p...)

	var resvec *Pooled[T]

	isFirst := true
	numMatches = 0

	for i := range svecs {

		svec := svecs[i]
		// In case the current vec is Rotated, we reduce it to a regular form
		// NB : this could use the pool.
		if rot, ok := svec.(*Rotated[T]); ok {
			svec = rotatedAsRegular(rot)
		}

		if pooled, ok := svec.(*Pooled[T]); ok {
			svec = &pooled.Regular
		}

		if reg, ok := svec.(*Regular[T]); ok {
			numMatches++
			// For the first one, we can save by just copying the result
			// Importantly, we do not need to assume that regRes is originally
			// zero.
			if isFirst {
				if hasPool {
					resvec = AllocFromPool(pool)
				} else {
					resvec = &Pooled[T]{Regular: make([]T, length)}
				}

				isFirst = false
				op.vecIntoTerm(resvec.Regular, *reg, coeffs[i])
				continue
			}

			op.vecIntoVec(resvec.Regular, *reg, coeffs[i])
		}
	}

	if numMatches == 0 {
		return nil, 0
	}

	return resvec, numMatches
}

func (r *Regular[T]) DeepCopy() SmartVector[T] {
	return NewRegular[T](vector.DeepCopy[T](*r))
}

// Converts a smart-vector into a normal vec. The implementation minimizes
// then number of copies.
func (r *Regular[T]) IntoRegVecSaveAlloc() []T {
	return (*r)[:]
}

type Pooled[T fmt.Stringer] struct {
	Regular[T]
	poolPtr *[]T
}

func AllocFromPool[T fmt.Stringer](pool mempool.MemPool[T]) *Pooled[T] {
	poolPtr := pool.Alloc()
	return &Pooled[T]{
		Regular: *poolPtr,
		poolPtr: poolPtr,
	}
}

func (p *Pooled[T]) Free(pool mempool.MemPool[T]) {
	if p.poolPtr != nil {
		pool.Free(p.poolPtr)
	}
	p.poolPtr = nil
	p.Regular = nil
}
