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
