package smartvectors

import (
	"fmt"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/mempool"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/vector"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// It's normal vector in a nutshell
type Regular []field.Element

// Instanstiate a new regular from a slice. Returns a pointer so that the result
// can be reused without referencing as a SmartVector.
func NewRegular(v []field.Element) *Regular {
	assertStrictPositiveLen(len(v))
	res := Regular(v)
	return &res
}

// Returns the length of the regular vector
func (r *Regular) Len() int { return len(*r) }

// Returns a particular element of the vector
func (r *Regular) Get(n int) field.Element { return (*r)[n] }

// Returns a subvector of the regular
func (r *Regular) SubVector(start, stop int) SmartVector {
	if start > stop {
		utils.Panic("Negative length are not allowed")
	}
	if start == stop {
		utils.Panic("Subvector of zero lengths are not allowed")
	}
	res := Regular((*r)[start:stop])
	return &res
}

// Rotates the vector into a new one
func (r *Regular) RotateRight(offset int) SmartVector {
	resSlice := make(Regular, r.Len())

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

func (r *Regular) WriteInSlice(s []field.Element) {
	assertHasLength(len(s), len(*r))
	copy(s, *r)
}

func (r *Regular) Pretty() string {
	return fmt.Sprintf("Regular[%v]", vector.Prettify(*r))
}

func processRegularOnly(op operator, svecs []SmartVector, coeffs []int, p ...*mempool.Pool) (result *Regular, numMatches int) {

	length := svecs[0].Len()

	pool, hasPool := mempool.ExtractCheckOptionalStrict(length, p...)

	var resvec []field.Element
	if hasPool {
		resvec = *pool.Alloc()
	} else {
		resvec = make([]field.Element, length)
	}

	isFirst := true
	numMatches = 0

	for i := range svecs {

		svec := svecs[i]
		// In case the current vec is Rotated, we reduce it to a regular form
		// NB : this could use the pool.
		if rot, ok := svec.(*Rotated); ok {
			svec = rotatedAsRegular(rot)
		}

		if reg, ok := svec.(*Regular); ok {
			numMatches++
			// For the first one, we can save by just copying the result
			// Importantly, we do not need to assume that regRes is originally
			// zero.
			if isFirst {
				isFirst = false
				op.vecIntoTerm(resvec, *reg, coeffs[i])
				continue
			}
			op.vecIntoVec(resvec, *reg, coeffs[i])
		}
	}

	if numMatches == 0 {
		if hasPool {
			// we did not need it, so we can give it back directly
			pool.Free(&resvec)
		}
		return nil, 0
	}

	res := Regular(resvec)
	return &res, numMatches
}

func (r *Regular) DeepCopy() SmartVector {
	return NewRegular(vector.DeepCopy(*r))
}

// Converts a smart-vector into a normal vec. The implementation minimizes
// then number of copies.
func (r *Regular) IntoRegVecSaveAlloc() []field.Element {
	return (*r)[:]
}
