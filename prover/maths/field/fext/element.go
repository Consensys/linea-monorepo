package fext

import (
	"fmt"
	"math/rand/v2"

	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// Embedding
type Element = extensions.E4

func NewElement(v1, v2, v3, v4 int64) Element {
	var z Element
	z.B0.A0.SetInt64(int64(v1))
	z.B0.A1.SetInt64(int64(v2))
	z.B1.A0.SetInt64(int64(v3))
	z.B1.A1.SetInt64(int64(v4))
	return z
}

// NewFromString only sets the first coordinate of the field extension
func NewFromString(s string) (res Element) {
	res.B0.A0 = field.NewFromString(s)
	return res
}

// var RootPowers = []int{1, 3}, v^2=u and u^2=3.
var RootPowers = []int{1, 3}

func BatchInvert(a []Element) []Element {
	return extensions.BatchInvertE4(a)
}

func PseudoRand(rng *rand.Rand) Element {

	result := new(Element).SetZero()
	result.B0.A0 = field.PseudoRand(rng)
	result.B0.A1 = field.PseudoRand(rng)
	result.B1.A0 = field.PseudoRand(rng)
	result.B1.A1 = field.PseudoRand(rng)

	return *result
}

/*
IsBase checks if the field extensionElement is a base element.
An Element is considered a base element if all coordinates are 0 except for the first one.
*/
func IsBase(z *Element) bool {
	if z.B0.A1[0] == 0 && z.B1.A0[0] == 0 && z.B1.A1[0] == 0 {
		return true
	} else {
		return false
	}

}

func GetBase(z *Element) (field.Element, error) {
	if IsBase(z) {
		return z.B0.A0, nil
	} else {
		return field.Zero(), fmt.Errorf("requested a base element but the field extension is not a wrapped field element")
	}
}
func AddByBase(z *Element, first *Element, second *field.Element) *Element {
	z.B0.A0.Add(&first.B0.A0, second)
	return z
}
func DivByBase(z *Element, first *Element, second *field.Element) *Element {
	z.B0.A0.Div(&first.B0.A0, second)
	z.B0.A1.Div(&first.B0.A1, second)
	z.B1.A0.Div(&first.B1.A0, second)
	z.B1.A1.Div(&first.B1.A1, second)
	return z
}
