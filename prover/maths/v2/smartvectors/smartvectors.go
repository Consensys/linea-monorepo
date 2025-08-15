package smartvectors

import (
	"math/rand/v2"

	field "github.com/consensys/linea-monorepo/prover/maths/v2/field"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// anyField is a union between field.Fr and field.Ext. It is used as type
// parameters for the implementations of [SmartVector].
type anyField interface {
	field.Fr | field.Ext
}

// SmartVector is an abstraction over vectors of field elements that can be
// optimized for structured vectors. For instance, if we have a vector of
// repeated elements we can use smartvectors.NewConstant(x, n) to represent it.
// This way instead of using n * sizeof(field.Element) memory it will only store
// the element once. Additionally, every operation performed on it will be
// sped up with dedicated algorithms.
//
// There are a few precautions to take when implementing or using smart-vectors
//   - constructing a zero-length smart-vector should be considered illegal. The
//     reason for such a restriction is tha t
//   - although the smart-vectors are not immutable, the user should refrain
//     mutating them after they are created as this may have unintended side
//     effects that are hard to track.
type SmartVector interface {
	// Len returns the length of the SmartVector
	Len() int
	// Get returns an entry of the SmartVector at particular position
	Get(int) field.Gen
	// GetPtr returns the element at position n. The returned pointer is
	// actually a copy of the element. So it can be modified without any side
	// effect.
	GetPtr(int) *field.Gen
	// SubVector returns a subvector of the [SmartVector]. It mirrors slice[Start:Stop]
	SubVector(int, int) SmartVector
	// RotateRight cyclically rotates the SmartVector
	RotateRight(int) SmartVector
	// DeepCopy returns a deep copy of the SmartVector
	DeepCopy() SmartVector
}

// WriteIntoSlice writes the smartvector into a slice
func WriteIntoSlice[T field.Gen | field.Fr | field.Ext](s_ []T, sv SmartVector) {

	if len(s_) != sv.Len() {
		utils.Panic("length mismatch: %v != %v", len(s_), sv.Len())
	}

	s := any(s_)

	switch {
	case utils.AreOfType[Regular[field.Ext], []field.Ext](sv, s):
		sv, s := utils.MustCastPair[Regular[field.Ext], []field.Ext](sv, s)
		copy(s, sv)

	case utils.AreOfType[Regular[field.Fr], []field.Ext](sv, s):
		sv, s := utils.MustCastPair[Regular[field.Fr], []field.Ext](sv, s)
		for i := range s {
			s[i] = sv[i].Lift()
		}

	case utils.AreOfType[Constant[field.Ext], []field.Ext](sv, s):
		sv, s := utils.MustCastPair[Constant[field.Ext], []field.Ext](sv, s)
		field.VecFill(s, sv.Value)

	case utils.AreOfType[Constant[field.Fr], []field.Ext](sv, s):
		sv, s := utils.MustCastPair[Constant[field.Fr], []field.Ext](sv, s)
		field.VecFill(s, sv.Value.Lift())

	case utils.AreOfType[PaddedCircularWindow[field.Ext], []field.Ext](sv, s):
		sv, s := utils.MustCastPair[PaddedCircularWindow[field.Ext], []field.Ext](sv, s)
		// Because of the wrap-around possibilty we don't use VecFill but there
		// might be a more efficient way to account for the wrap-around.
		for i := range sv.Window_ {
			pos := utils.PositiveMod(i+sv.Offset_, sv.TotLen_)
			s[pos] = sv.Window_[i]
		}
		for i := len(sv.Window_); i < sv.TotLen_; i++ {
			pos := utils.PositiveMod(i+sv.Offset_, sv.TotLen_)
			s[pos] = sv.PaddingVal_
		}

	case utils.AreOfType[PaddedCircularWindow[field.Fr], []field.Ext](sv, s):
		sv, s := utils.MustCastPair[PaddedCircularWindow[field.Fr], []field.Ext](sv, s)
		// Because of the wrap-around possibilty we don't use VecFill but there
		// might be a more efficient way to account for the wrap-around.
		for i := range sv.Window_ {
			pos := utils.PositiveMod(i+sv.Offset_, sv.TotLen_)
			s[pos] = sv.Window_[i].Lift()
		}
		for i := len(sv.Window_); i < sv.TotLen_; i++ {
			pos := utils.PositiveMod(i+sv.Offset_, sv.TotLen_)
			s[pos] = sv.PaddingVal_.Lift()
		}
	}
}

// SmartVectorToSlice converts a smartvector into a slice of the requested type.
// The function deep-copies the smartvector and will never save allocations.
func SmartVectorToSlice[T field.Gen | field.Fr | field.Ext](sv SmartVector) []T {
	panic("to be implemented")
}

// Rand creates a base vector with random entries. Used for testing. Should not be
// used to generate secrets. Not reproducible.
func Rand[T field.Fr | field.Ext](n int) SmartVector {
	panic("to be implemented")
}

// Rand creates a vector with random entries. Used for testing. Should not be
// used to generate secrets. Takes a math.Rand as input for reproducibility
// math
func PseudoRand[T field.Fr | field.Ext](rng *rand.Rand, n int) SmartVector {
	panic("to be implemented")
}

// FromInts creates a vector from a slice of integers
func FromInts(ints []int) SmartVector {
	panic("to be implemented")
}

// FromIntsQuads returns a vector of field elements from a slice of [4]integers.
func FromIntsQuads(ints [][4]int) SmartVector {
	panic("to be implemented")
}

// PaddingVal returns if the smart-vector has padding and the padding value as
// a Gen.
func PaddingVal(v SmartVector) bool {
	panic("to be implemented")
}
