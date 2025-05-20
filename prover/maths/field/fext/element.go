package fext

import (
	"math/rand/v2"

	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/linea-monorepo/prover/maths/field"
)

// Embedding
type Element = extensions.E4

func NewElement(v1 uint32, v2 uint32, v3 uint32, v4 uint32) Element {
	var z Element
	z.B0.A0 = field.Element{v1}
	z.B0.A1 = field.Element{v2}
	z.B1.A0 = field.Element{v3}
	z.B1.A1 = field.Element{v4}
	return z
}

// NewFromString only sets the first coordinate of the field extension
func NewFromString(s string) (res Element) {
	res.B0.A0 = field.NewFromString(s)
	return res
}

// var RootPowers = []int{1, -11} // ??
var RootPowers = []int{1, 3} // ??

func BatchInvertE4(a []Element) []Element {
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
