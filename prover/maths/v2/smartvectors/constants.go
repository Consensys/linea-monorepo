package smartvectors

import (
	"github.com/consensys/linea-monorepo/prover/maths/v2/field"
)

// Constant represents a constant vector obtained by repeated "length" time the
// same value
type Constant[T anyField] struct {
	Value  T
	Length int
}

// NewConstant returns a new "Constant" smart-vector
func NewConstant[T anyField](val T, length int) *Constant[T] {
	assertValidLen(length)
	return &Constant[T]{Value: val, Length: length}
}

// NewConstantGen is as [NewConstant] and takes a [field.Gen] as input
func NewConstantGen(val field.Gen, length int) SmartVector {
	if val, ok := val.AsFr(); ok {
		return NewConstant(val, length)
	}
	return NewConstant(val.AsExt(), length)
}

// Return the length of the smart-vector
func (c *Constant[T]) Len() int { return c.Length }

// Get always returns the value of the constant. The function still panics if
// length is out of bound.
func (r *Constant[T]) Get(n int) field.Gen {
	assertInBound(n, r.Length)
	return field.NewGen(r.Value)
}

// GetPtr delegates to [Get] and returns the address of the result
func (r *Constant[T]) GetPtr(n int) *field.Gen {
	v := r.Get(n)
	return &v
}

// SubVector returns a smaller length version of the [Constant].
func (c *Constant[T]) SubVector(start, stop int) SmartVector {
	assertValidRange(start, stop)
	assertInBound(start, c.Length)
	assertInBound(stop, c.Length+1) // The +1 is because we accept if "Stop = length"
	return NewConstant[T](c.Value, stop-start)
}

// RotateRight is ineffective over a constant smartvector. We just return the
// receiver.
func (c *Constant[T]) RotateRight(int) SmartVector {
	return NewConstant(c.Value, c.Length)
}

// DeepCopy just reconstruct the same smart-vector
func (c *Constant[T]) DeepCopy() SmartVector {
	return NewConstant(c.Value, c.Length)
}
