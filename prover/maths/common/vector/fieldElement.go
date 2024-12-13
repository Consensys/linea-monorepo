package vector

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"math/rand"
)

type fieldElement interface {
	SetRandom() (*fieldElement, error)
	String() string
}

func PseudoRandElement[T fieldElement](rng *rand.Rand) T {
	// Generic function to return a random value of type int, bool, or float64
	var result T
	switch any(result).(type) {
	case fr.Element:
		return any(field.PseudoRand(rng)).(T)
	case fext.Element:
		return any(fext.PseudoRand(rng)).(T)
	default:
		panic("unsupported type")
	}
	return result
}
