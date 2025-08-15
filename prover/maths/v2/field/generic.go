package field

import (
	"math/big"

	"github.com/consensys/linea-monorepo/prover/utils"
)

// Gen is a union type of Fr and Ext. It is used externally a
// little bit everywhere to represent either of these types. The zero value
// of this struct represents the zero element of the base field.
type Gen struct {
	IsExt_ bool
	Base_  Fr
	Ext_   Ext
	_      [0]func() // makes the structure non comparable with the == operator
}

// NewGen returns a new generic field from an any
func NewGen(x any) Gen {
	return *new(Gen).Set(x)
}

func (z *Gen) Set(x any) *Gen {

	z.assertWellFormed()
	defer z.maybeSimplify()
	defer z.assertWellFormed()

	// Here we can't relie on unwrapAny because the function also accepts non
	// field inputs and does the conversion.

	switch x := x.(type) {
	case Fr:
		z.IsExt_ = false
		z.Base_ = x
		z.Ext_.SetZero()
	case Ext:
		z.Base_.SetZero()
		z.Ext_ = x
		z.IsExt_ = true
	case Gen:
		x.assertWellFormed()
		z.IsExt_ = x.IsExt_
		z.Base_ = x.Base_
		z.Ext_ = x.Ext_
	case *Fr:
		z.IsExt_ = false
		z.Base_ = *x
		z.Ext_.SetZero()
	case *Ext:
		z.Base_.SetZero()
		z.Ext_ = *x
		z.IsExt_ = true
	case *Gen:
		x.assertWellFormed()
		z.IsExt_ = x.IsExt_
		z.Base_ = x.Base_
		z.Ext_ = x.Ext_
	case uint64:
		z.IsExt_ = false
		z.Base_.SetUint64(x)
		z.Ext_.SetZero()
	case int64:
		z.IsExt_ = false
		z.Base_.SetInt64(x)
		z.Ext_.SetZero()
	default:
		utils.Panic("can't set Gen with %T", x)
	}

	return z
}

func (z *Gen) IsBase() bool {
	z.assertWellFormed()
	return !z.IsExt_
}

func (z *Gen) IsExt() bool {
	z.assertWellFormed()
	return z.IsExt_
}

func (z *Gen) AsFr() (Fr, bool) {
	z.assertWellFormed()
	z.maybeSimplify()
	// The simplification routine
	if z.IsExt_ {
		if z.Ext_.IsFr() {
			panic("the simplification routine should have simplified")
		}
		return Fr{}, false
	}
	return z.Base_, true
}

func (z *Gen) AsExt() Ext {
	z.assertWellFormed()
	if !z.IsExt_ {
		return z.Base_.Lift()
	}
	return z.Ext_
}

func (z *Gen) Underlying() any {
	z.assertWellFormed()
	z.maybeSimplify()
	if z.IsExt_ {
		return &z.Ext_
	}
	return &z.Base_
}

func (z *Gen) String() string {
	return z.Underlying().(interface{ String() string }).String()
}

func (z *Gen) Bytes() []byte {
	return z.Underlying().(interface{ Bytes() []byte }).Bytes()
}

func GenOne() Gen {
	return Gen{IsExt_: false, Base_: One[Fr]()}
}

func GenZero() Gen {
	return Gen{}
}

func (z *Gen) SetZero() *Gen {
	z.IsExt_ = false
	z.Base_.SetZero()
	z.Ext_.SetZero()
	return z
}

func (z *Gen) SetOne() *Gen {
	z.IsExt_ = false
	z.Base_.SetOne()
	z.Ext_.SetZero()
	return z
}

func (z *Gen) IsZero() bool {
	return z.Base_.IsZero() && z.Ext_.IsZero()
}

func (z *Gen) IsOne() bool {
	z.assertWellFormed()
	return z.Base_.IsOne() || z.Ext_.IsOne()
}

func (z *Gen) IsEqual(x any) bool {

	// TODO: could be optimized but that makes the comparison much simpler if
	// if we don't have to deal with the fact that the Gens can be Ext but still
	// represents an Fr value.
	z.maybeSimplify()

	x = unwrapFieldAny(x)

	switch x := x.(type) {

	case *Fr:
		return z.IsBase() && z.Base_ == *x

	case *Ext:
		if x.IsFr() {
			return z.IsBase() && z.Base_ == Fr(x.B0.A0)
		}
		return z.IsExt() && z.Ext_ == *x

	default:
		return false
	}

}

func (z *Gen) Neg(x any) *Gen {

	defer z.assertWellFormed()

	x = unwrapFieldAny(x)

	switch x := x.(type) {
	case *Fr:
		return z.negFr(x)
	case *Ext:
		return z.negExt(x)
	default:
		utils.Panic("can't negate %T", x)
	}

	panic("unreachable")
}

func (z *Gen) Inverse(x any) *Gen {

	defer z.assertWellFormed()
	x = unwrapFieldAny(x)

	switch x := x.(type) {
	case *Fr:
		return z.inverseFr(x)
	case *Ext:
		return z.inverseExt(x)
	default:
		utils.Panic("can't negate %T", x)
	}

	panic("unreachable")
}

func (z *Gen) Square(x any) *Gen {
	return z.Mul(x, x)
}

func (z *Gen) Mul(x, y any) *Gen {

	z.Set(x)
	y = unwrapFieldAny(y)

	switch y := y.(type) {
	case *Fr:
		z.mulAssignFr(y)
	case *Ext:
		z.mulAssignExt(y)
	default:
		utils.Panic("can't multiply %T with %T", x, y)
	}

	return z
}

func (z *Gen) Exp(x any, n int) *Gen {

	z.assertWellFormed()
	z.Set(x)
	z.assertWellFormed()
	z.maybeSimplify()

	if z.IsExt_ {
		z.Ext_.ExpToInt(z.Ext_, n)
	} else {
		z.Base_.Exp(z.Base_, big.NewInt(int64(n)))
	}

	return z
}

func (z *Gen) Add(x, y any) *Gen {

	z.Set(x)
	y = unwrapFieldAny(y)

	switch y := y.(type) {
	case *Fr:
		z.addAssignFr(y)
	case *Ext:
		z.addAssignExt(y)
	default:
		utils.Panic("can't multiply %T with %T", x, y)
	}

	return z
}

func (z *Gen) Sub(x, y any) *Gen {

	z.Set(x)
	y = unwrapFieldAny(y)

	switch y := y.(type) {
	case *Fr:
		z.subAssignFr(y)
	case *Ext:
		z.subAssignExt(y)
	default:
		utils.Panic("can't multiply %T with %T", x, y)
	}

	return z
}

func (z *Gen) Div(x, y any) *Gen {

	z.Set(x)
	y = unwrapFieldAny(y)

	switch y := y.(type) {
	case *Fr:
		z.divAssignFr(y)
	case *Ext:
		z.divAssignExt(y)
	default:
		utils.Panic("can't multiply %T with %T", x, y)
	}

	return z
}

func (z *Gen) negFr(x *Fr) *Gen {
	z.Base_.Neg(x)
	z.Ext_.SetZero()
	z.IsExt_ = false
	return z
}

func (z *Gen) negExt(x *Ext) *Gen {
	z.Ext_.Neg(x)
	z.Base_.SetZero()
	z.IsExt_ = true
	return z
}

func (z *Gen) inverseFr(x *Fr) *Gen {
	z.Base_.Inverse(x)
	z.Ext_.SetZero()
	z.IsExt_ = false
	return z
}

func (z *Gen) inverseExt(x *Ext) *Gen {
	z.Ext_.Inverse(x)
	z.Base_.SetZero()
	z.IsExt_ = true
	return z
}

func (z *Gen) mulAssignFr(x *Fr) {
	z.assertWellFormed()
	defer z.assertWellFormed()
	if z.IsExt_ {
		z.Ext_.MulByElement(&z.Ext_, x)
	} else {
		z.Base_.Mul(&z.Base_, x)
	}
}

func (z *Gen) mulAssignExt(x *Ext) {
	z.assertWellFormed()
	defer z.assertWellFormed()
	defer z.maybeSimplify()
	if z.IsExt_ {
		z.Ext_.Mul(&z.Ext_, x)
	} else {
		z.Ext_.MulByElement(x, &z.Base_)
		z.Base_.SetZero()
		z.IsExt_ = true
	}
}

func (z *Gen) addAssignFr(x *Fr) {
	z.assertWellFormed()
	defer z.assertWellFormed()
	if z.IsExt_ {
		z.Ext_.AddFr(&z.Ext_, x)
	} else {
		z.Base_.Add(&z.Base_, x)
	}
}

func (z *Gen) addAssignExt(x *Ext) {
	z.assertWellFormed()
	defer z.assertWellFormed()
	defer z.maybeSimplify()
	if z.IsExt_ {
		z.Ext_.Add(&z.Ext_, x)
	} else {
		z.Ext_.AddFr(x, &z.Base_)
		z.Base_.SetZero()
		z.IsExt_ = true
	}
}

func (z *Gen) subAssignFr(x *Fr) {
	if z.IsExt_ {
		z.Ext_.SubExtByFr(&z.Ext_, x)
	} else {
		z.Base_.Sub(&z.Base_, x)
	}
}

func (z *Gen) subAssignExt(x *Ext) {
	z.assertWellFormed()
	defer z.assertWellFormed()
	defer z.maybeSimplify()
	if z.IsExt_ {
		z.Ext_.Sub(&z.Ext_, x)
	} else {
		z.Ext_.SubFrByExt(&z.Base_, x)
		z.Base_.SetZero()
		z.IsExt_ = true
	}
}

func (z *Gen) divAssignFr(x *Fr) {
	if z.IsExt_ {
		z.Ext_.DivExtByFr(&z.Ext_, x)
	} else {
		z.Base_.Div(&z.Base_, x)
	}
}

func (z *Gen) divAssignExt(x *Ext) {
	z.assertWellFormed()
	defer z.assertWellFormed()
	defer z.maybeSimplify()
	if z.IsExt_ {
		z.Ext_.Div(&z.Ext_, x)
	} else {
		z.Ext_.DivFrByExt(&z.Base_, x)
		z.Base_.SetZero()
		z.IsExt_ = true
	}
}

// assertWellFormed ensures that the field is well-formed
func (z *Gen) assertWellFormed() {
	if z.IsExt_ && !z.Base_.IsZero() {
		panic("can't have a non-zero base field element in an extension field")
	}

	if !z.IsExt_ && !z.Ext_.IsZero() {
		panic("can't have a non-zero extension field element in a base field")
	}
}

// maybeSimplify tries to simplify the field into a base field element if the
// current instance is assigned to an extension field with zero as imaginary
// parts.
func (z *Gen) maybeSimplify() {
	defer z.assertWellFormed()
	if z.IsExt_ && z.Ext_.IsFr() {
		z.IsExt_ = false
		z.Base_ = Fr(z.Ext_.B0.A0)
		z.Ext_.SetZero()
	}
}

// unwrapFieldAny takes a value of any either [*Fr, Fr, *Ext, Ext, Gen, *Gen] and
// returns a value of either *Fr or *Ext depending on the instance. The function
// panics if any other case is provided.
func unwrapFieldAny(x any) any {
	switch x := x.(type) {
	case *Fr:
		return x
	case Fr:
		return &x
	case *Ext:
		return x
	case Ext:
		return &x
	case *Gen:
		return x.Underlying()
	case Gen:
		return x.Underlying()
	default:
		panic("unsupported type")
	}
}
