package field

import (
	"math/bits"

	"github.com/consensys/gnark-crypto/field/koalabear"
)

type Element = koalabear.Element

// NewElement constructs a new field element corresponding to an integer.
var NewElement = koalabear.NewElement
var BatchInvert = koalabear.BatchInvert

// Zero returns the zero field element
func Zero() Element {
	var res Element
	return res
}

// One returns the one field element
func One() Element {
	var res Element
	res.SetUint64(1)
	return res
}

// NewFromString constructs a new field element from a string. The rules to
// determine how the string is casted into a field elements are the one of
// [fr.Element.SetString]
func NewFromString(s string) (res Element) {
	_, err := res.SetString(s)
	if err != nil {
		panic(err)
	}
	return res
}

// ExpToInt sets z to x**k
func ExpToInt(z *Element, x Element, k int) *Element {
	if k == 0 {
		return z.SetOne()
	}

	if k < 0 {
		x.Inverse(&x)
		k = -k
	}

	z.Set(&x)

	for i := bits.Len(uint(k)) - 2; i >= 0; i-- {
		z.Square(z)
		if (k>>i)&1 == 1 {
			z.Mul(z, &x)
		}
	}

	return z
}
