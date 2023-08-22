package smartvectors

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

// Represents a rotated version of a regular smartvector
type Rotated struct {
	v      Regular
	offset int
}

// Construct a new rotated
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

func (*Rotated) AddRef() {}
func (*Rotated) DecRef() {}
func (*Rotated) Drop()   {}

func (r *Rotated) WriteInSlice(s []field.Element) {
	// the casting is a mechanism to prevent against infinity loop
	// in case we decide that subvectors of rotated are no longer
	// always regular.
	res := rotatedAsRegular(r)
	res.WriteInSlice(s)

}

func (r *Rotated) Pretty() string {
	v := &r.v
	return fmt.Sprintf("Rotated[%v, %v]", v.Pretty(), r.offset)
}

func rotatedAsRegular(r *Rotated) *Regular {
	return r.SubVector(0, r.Len()).(*Regular)
}

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
