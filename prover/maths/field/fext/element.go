package fext

import (
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

// var RootPowers = []int{1, -11} // ??
var RootPowers = []int{1, 3} // ??

func BatchInvertE4(a []Element) []Element {
	return extensions.BatchInvertE4(a)
}
