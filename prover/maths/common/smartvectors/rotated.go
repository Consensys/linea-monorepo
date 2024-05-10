package smartvectors

import (
	"fmt"

	"github.com/consensys/zkevm-monorepo/prover/maths/common/vector"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// Rotated represents a rotated version of a regular smartvector and also
// implements the [SmartVector] interface. Rotated have a very niche use-case
// in the repository as they are used to help saving FFT operations in the
// [github.com/consensys/zkevm-monorepo/prover/protocol/compiler/arithmetic.CompileGlobal]
// compiler when the coset evaluation is done over a cyclic rotation of a
// smart-vector.
//
// Rotated works by abstractly storing the offset and only applying the rotation
// when the vector is written or sub-vectored. This makes rotations essentially
// free.
type Rotated struct {
	v      Regular
	offset int
}

// NewRotated constructs a new Rotated
func NewRotated(reg Regular, offset int) *Rotated {

	// empty vector
	if len(reg) == 0 {
		utils.Panic("got an empty vector")
	}

	// offset larger than the vector itself
	if offset > len(reg) {
		utils.Panic("len %v, offset %v", len(reg), offset)
	}

	return &Rotated{
		v: reg, offset: offset,
	}
}

// Returns the lenght of the vector
func (r *Rotated) Len() int {
	return r.v.Len()
}

// Returns a particular element of the vector
func (r *Rotated) Get(n int) field.Element {
	return r.v.Get(utils.PositiveMod(n+r.offset, r.Len()))
}

// Returns a particular element. The subvector is taken at indices
// [start, stop). (stop being excluded from the span)
func (r *Rotated) SubVector(start, stop int) SmartVector {
	res := make([]field.Element, stop-start)
	size := r.Len()
	spanSize := stop - start

	// checking
	if stop <= start {
		utils.Panic("the start %v >= stop %v", start, stop)
	}

	// boundary checks
	if start < 0 {
		utils.Panic("the start value was negative %v", start)
	}

	if stop > size {
		utils.Panic("the stop is OOO : %v (the length is %v)", stop, size)
	}

	// normalize the offset to something positive [0: size)
	startWithOffsetClean := utils.PositiveMod(start+r.offset, size)

	// NB: we may need to construct the res in several steps
	// in case
	copy(res, r.v[startWithOffsetClean:utils.Min(size, startWithOffsetClean+spanSize)])

	// If this is negative of zero, it means the first copy already copied
	// everything we needed to copy
	howManyElementLeftToCopy := startWithOffsetClean + spanSize - size
	howManyAlreadyCopied := spanSize - howManyElementLeftToCopy
	if howManyElementLeftToCopy <= 0 {
		ret := Regular(res)
		return &ret
	}

	// if necessary perform a second
	copy(res[howManyAlreadyCopied:], r.v[:howManyElementLeftToCopy])
	ret := Regular(res)
	return &ret
}

// Rotates the vector into a new one
func (r *Rotated) RotateRight(offset int) SmartVector {
	return &Rotated{
		v:      vector.DeepCopy(r.v),
		offset: r.offset + offset,
	}
}

func (r *Rotated) DeepCopy() SmartVector {
	return NewRotated(vector.DeepCopy(r.v), r.offset)
}

func (r *Rotated) WriteInSlice(s []field.Element) {
	res := rotatedAsRegular(r)
	res.WriteInSlice(s)
}

func (r *Rotated) Pretty() string {
	v := &r.v
	return fmt.Sprintf("Rotated[%v, %v]", v.Pretty(), r.offset)
}

// rotatedAsRegular converts a [Rotated] into a [Regular] by effecting the
// symbolic shifting operation. The function allocates the result.
func rotatedAsRegular(r *Rotated) *Regular {
	return r.SubVector(0, r.Len()).(*Regular)
}

func (r *Rotated) IntoRegVecSaveAlloc() []field.Element {
	return *rotatedAsRegular(r)
}

// SoftRotate converts v into a [SmartVector] representing the same
// [SmartVector]. The function tries to not reallocate the result. This means
// that changing the v can subsequently affects the result of this function.
func SoftRotate(v SmartVector, offset int) SmartVector {

	switch casted := v.(type) {
	case *Regular:
		return NewRotated(*casted, offset)
	case *Rotated:
		return NewRotated(casted.v, utils.PositiveMod(offset+casted.offset, v.Len()))
	case *PaddedCircularWindow:
		return NewPaddedCircularWindow(
			casted.window,
			casted.paddingVal,
			utils.PositiveMod(casted.offset+offset, casted.Len()),
			casted.Len(),
		)
	case *Constant:
		// It's a constant so it does not need to be rotated
		return v
	default:
		utils.Panic("unknown type %T", v)
	}

	panic("unreachable")

}
