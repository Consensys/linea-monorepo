package column

import (
	"math/big"
	"reflect"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
)

// GnarkDeriveEvaluationPoint mirrors [DeriveEvaluationPoint] but in a gnark
// circuit
func GnarkDeriveEvaluationPoint[T zk.Element](
	api frontend.API, h ifaces.Column[T], upstream string,
	cachedXs collection.Mapping[string, T],
	x T,
) (xRes []T) {

	if !h.IsComposite() {
		// Just return x and cache it if necessary
		newUpstream := appendNodeToUpstream(upstream, h)
		// Store in the cache if necessary
		if !cachedXs.Exists(newUpstream) {
			// Else register the result in the cache
			cachedXs.InsertNew(newUpstream, x)
		}
		return []T{x}
	}

	apiGen, err := zk.NewApi[T](api)
	if err != nil {

	}

	switch inner := h.(type) {
	case Shifted[T]:
		newUpstream := appendNodeToUpstream(upstream, inner)
		var derivedX T
		// Early return if the result is cached
		if cachedXs.Exists(newUpstream) {
			derivedX = cachedXs.MustGet(newUpstream)
		} else {
			// If not, compute the shift on x and cache the result
			n := h.Size()
			generator, err := fft.Generator(uint64(n))
			if err != nil {
				panic(err)
			}
			// omegaN := T(generator)
			omegaN := apiGen.FromKoalabear(generator)
			omegaN = field.Exp(apiGen, *omegaN, big.NewInt(int64(inner.Offset)))
			derivedX = *apiGen.Mul(&x, omegaN)
			cachedXs.InsertNew(newUpstream, derivedX)
		}
		return GnarkDeriveEvaluationPoint(api, inner.Parent, newUpstream, cachedXs, derivedX)

	default:
		utils.Panic("unexpected type %v", reflect.TypeOf(inner))
	}
	panic("unreachable")
}

// GnarkVerifyYConsistency does the same as [VerifyYConsistency] but in a gnark
// circuit.
func GnarkVerifyYConsistency[T zk.Element](
	api frontend.API, h ifaces.Column[T], upstream string,
	cachedXs collection.Mapping[string, T],
	finalYs collection.Mapping[string, T],
) (y T) {

	if !h.IsComposite() {
		// Get the Y from the map. An absence from this map is unexpected at
		// this level.
		newUpstream := appendNodeToUpstream(upstream, h)
		res := finalYs.MustGet(DerivedYRepr(newUpstream, h))
		return res
	}

	switch inner := h.(type) {
	case Shifted[T]:
		newUpstream := appendNodeToUpstream(upstream, inner)
		res := GnarkVerifyYConsistency(api, inner.Parent, newUpstream, cachedXs, finalYs)
		// No need to test the error, because we would return it alonside the
		// nil result anyway.
		return res

	default:
		utils.Panic("unexpected type %v", reflect.TypeOf(inner))
	}
	panic("unreachable")
}
