package field

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
)

type Element = koalabear.Element

// NewElement constructs a new field element corresponding to an integer.
var NewElement = koalabear.NewElement

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
