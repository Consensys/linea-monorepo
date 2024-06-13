package smartvectors

import (
	"fmt"

	"github.com/consensys/zkevm-monorepo/prover/maths/field"
)

// A constant vector is a vector obtained by repeated "length" time the same value
type Constant struct {
	val    field.Element
	length int
}

// Construct a new "Constant" smart-vector
func NewConstant(val field.Element, length int) *Constant {
	return &Constant{val: val, length: length}
}

// Return the length of the smart-vector
func (c *Constant) Len() int { return c.length }

// Returns an entry of the constant
func (c *Constant) Get(int) field.Element { return c.val }

// Returns a subvector
func (c *Constant) SubVector(start, stop int) SmartVector {
	assertCorrectBound(start, c.length)
	// The +1 is because we accept if "stop = length"
	assertCorrectBound(stop, c.length+1)
	return NewConstant(c.val, stop-start)
}

// Returns a rotated version of the slice
func (c *Constant) RotateRight(int) SmartVector {
	return NewConstant(c.val, c.length)
}

// Write the constant vector in a slice
func (c *Constant) WriteInSlice(s []field.Element) {
	assertHasLength(len(s), c.Len())
	for i := range s {
		s[i] = c.val
	}
}

func (c *Constant) Val() field.Element {
	return c.val
}

func (c *Constant) Pretty() string {
	return fmt.Sprintf("Constant[%v;%v]", c.val.String(), c.length)
}

func (c *Constant) DeepCopy() SmartVector {
	return NewConstant(c.val, c.length)
}

func (c *Constant) IntoRegVecSaveAlloc() []field.Element {
	return IntoRegVec(c)
}
