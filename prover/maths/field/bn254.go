//go:build !goldilocks

package field

import (
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// Flag indicating whether we are using goldilocks or not
const USING_GOLDILOCKS = false

// Type alias to make it easy to switch
type Element = fr.Element

// Modulus of the
var Modulus = fr.Modulus

// Function alias for the butterfly
var Butterfly = fr.Butterfly

var One = fr.One

var NewElement = fr.NewElement

// Convenient function that returns zero
func Zero() Element {
	var res Element
	return res
}

const Bits = fr.Bits
const Bytes = fr.Bytes

var BatchInvert = fr.BatchInvert

const RootOfUnity string = "19103219067921713944291392827692070036145651957329286315305642004821462161904"
const RootOrUnityOrder uint64 = 28
const MultiplicativeGen uint64 = 5

// Sugar for string constructors
func NewFromString(s string) (res Element) {
	_, err := res.SetString(s)
	if err != nil {
		utils.Panic("Invalid string encoding %v", s)
	}
	return res
}

// Batch invert but reuses a given buffer
// BatchInvert returns a new slice with every element inverted.
// Uses Montgomery batch inversion trick
func BatchInvertWithBuffer(res, a []Element) []Element {

	if len(a) == 0 {
		utils.Panic("empty batch invert")
	}

	if len(res) != len(a) {
		utils.Panic("the buffer has the wrong size (%v) should be %v", len(res), len(a))
	}

	zeroes := make([]bool, len(a))
	accumulator := One()

	for i := 0; i < len(a); i++ {
		if a[i].IsZero() {
			zeroes[i] = true
			res[i] = Zero() // zeroes so that we can track it thereafter
		}
		res[i] = accumulator
		accumulator.Mul(&accumulator, &a[i])
	}

	accumulator.Inverse(&accumulator)

	for i := len(a) - 1; i >= 0; i-- {
		if zeroes[i] {
			continue
		}
		res[i].Mul(&res[i], &accumulator)
		accumulator.Mul(&accumulator, &a[i])
	}

	return res
}
