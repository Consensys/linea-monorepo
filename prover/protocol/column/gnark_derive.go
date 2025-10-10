package column

import (
	"reflect"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
	"github.com/consensys/linea-monorepo/prover/utils/gnarkutil"
)

// GnarkDeriveEvaluationPoint mirrors [DeriveEvaluationPoint] but in a gnark
// circuit
func GnarkDeriveEvaluationPoint(
	api frontend.API, h ifaces.Column, upstream string,
	cachedXs collection.Mapping[string, zk.WrappedVariable],
	x zk.WrappedVariable,
) (xRes []zk.WrappedVariable) {

	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		panic(err)
	}

	if !h.IsComposite() {
		// Just return x and cache it if necessary
		newUpstream := appendNodeToUpstream(upstream, h)
		// Store in the cache if necessary
		if !cachedXs.Exists(newUpstream) {
			// Else register the result in the cache
			cachedXs.InsertNew(newUpstream, x)
		}
		return []zk.WrappedVariable{x}
	}

	switch inner := h.(type) {
	case Shifted:
		newUpstream := appendNodeToUpstream(upstream, inner)
		var derivedX zk.WrappedVariable
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
			omegaN := zk.ValueOf(generator)
			omegaN = gnarkutil.Exp(api, omegaN, inner.Offset)
			derivedX = *apiGen.Mul(&x, &omegaN)
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
func GnarkVerifyYConsistency(
	api frontend.API, h ifaces.Column, upstream string,
	cachedXs collection.Mapping[string, zk.WrappedVariable],
	finalYs collection.Mapping[string, zk.WrappedVariable],
) (y zk.WrappedVariable) {

	if !h.IsComposite() {
		// Get the Y from the map. An absence from this map is unexpected at
		// this level.
		newUpstream := appendNodeToUpstream(upstream, h)
		res := finalYs.MustGet(DerivedYRepr(newUpstream, h))
		return res
	}

	switch inner := h.(type) {
	case Shifted:
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
