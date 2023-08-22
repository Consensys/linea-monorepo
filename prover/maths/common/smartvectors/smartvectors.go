package smartvectors

import (
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark/frontend"
)

type SmartVector interface {
	// Returns the length of the smart-vector
	Len() int
	// Returns an entry for the smart-vector
	Get(int) field.Element
	// Returns a subvector v[start:stop] of the slice
	SubVector(int, int) SmartVector
	RotateRight(int) SmartVector
	WriteInSlice([]field.Element)
	Pretty() string
	DeepCopy() SmartVector
	AddRef()
	DecRef()
	Drop()
}

// Returns a newly allocated smart-vector
func AllocateRegular(n int) SmartVector {
	return NewRegular(make([]field.Element, n))
}

// Copy into a smart-vector, will panic if into is not a regular
// Mainly used as a sugar for refactoring
func Copy(into *SmartVector, x SmartVector) {
	*into = x.DeepCopy()
}

// Usefull for testing. Creates a vector with random entries
func Rand(n int) SmartVector {
	v := vector.Rand(n)
	return NewRegular(v)
}

// Returns a witness from a explicit litteral assignement
func ForTest(xs ...int) SmartVector {
	return NewRegular(vector.ForTest(xs...))
}

// Converts a smart-vector into a normal vec
func IntoRegVec(s SmartVector) []field.Element {
	res := make([]field.Element, s.Len())
	s.WriteInSlice(res)
	return res
}

// Converts a smart-vector into a gnark assignment
func IntoGnarkAssignment(sv SmartVector) []frontend.Variable {
	res := make([]frontend.Variable, sv.Len())
	for i := range res {
		res[i] = sv.Get(i)
	}
	return res
}

// Creates a new padded vector (padded on the left)
func LeftPadded(v []field.Element, padding field.Element, targetLen int) SmartVector {

	if len(v) > targetLen {
		utils.Panic("target length %v must be less than %v", len(v), targetLen)
	}

	if len(v) == targetLen {
		return NewRegular(v)
	}

	if len(v) == 0 {
		return NewConstant(padding, targetLen)
	}

	return NewPaddedCircularWindow(v, padding, targetLen-len(v), targetLen)
}

// Creates a new vector (padded on the right)
func RightPadded(v []field.Element, padding field.Element, targetLen int) SmartVector {

	if len(v) > targetLen {
		utils.Panic("target length %v must be less than %v", len(v), targetLen)
	}

	if len(v) == targetLen {
		return NewRegular(v)
	}

	if len(v) == 0 {
		return NewConstant(padding, targetLen)
	}

	return NewPaddedCircularWindow(v, padding, 0, targetLen)
}

// Creates a new padded vector (padded on the left)
func LeftZeroPadded(v []field.Element, targetLen int) SmartVector {
	return LeftPadded(v, field.Zero(), targetLen)
}

// Creates a new vector (padded on the right)
func RightZeroPadded(v []field.Element, targetLen int) SmartVector {
	return RightPadded(v, field.Zero(), targetLen)
}
