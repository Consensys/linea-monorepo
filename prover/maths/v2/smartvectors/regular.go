package smartvectors

import (
	"github.com/consensys/linea-monorepo/prover/maths/v2/field"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// Regular represents a slice of T and implements the [SmartVector] interface.
type Regular[T anyField] []T

// NewRegular returns a new regular vector. It will panic if the slice is empty
func NewRegular[T anyField](v []T) *Regular[T] {
	assertValidLen(len(v))
	res := Regular[T](v)
	return &res
}

// Len returns the length of the regular vector
func (r Regular[T]) Len() int { return len(r) }

// Get returns a particular element of the vector
func (r Regular[T]) Get(n int) field.Gen {
	assertInBound(n, r.Len())
	x := r[n]
	return field.NewGen(x)
}

// GetPtr returns the element at position n. The returned pointer is actually a
// copy of the element. So it can be modified without any side effect.
func (r Regular[T]) GetPtr(n int) *field.Gen {
	v := r.Get(n)
	return &v
}

// Returns a subvector of the regular. The result is deep-copied.
func (r Regular[T]) SubVector(start, stop int) SmartVector {
	assertValidRange(start, stop)
	return append(Regular[T]{}, r[start:stop]...)
}

// Rotates the vector into a new one. The function allocates its result.
func (r Regular[T]) RotateRight(offset int) SmartVector {
	resSlice := make(Regular[T], r.Len())

	if offset == 0 {
		copy(resSlice, r)
		return resSlice
	}

	if offset < 0 || offset > len(r) {
		offset = utils.PositiveMod(offset, len(r))
	}

	// v and w may be the same vector thus we should use a
	// separate leftover buffer for temporary memory buffers.
	cutAt := len(r) - offset
	copy(resSlice[offset:], r[:cutAt])
	copy(resSlice[:offset], r[cutAt:])
	return resSlice
}

// DeepCopy returns a deep copy of the vector
func (r Regular[T]) DeepCopy() SmartVector {
	return append(Regular[T]{}, r...)
}
