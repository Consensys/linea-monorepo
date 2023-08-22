package column

import (
	"fmt"
	"reflect"

	"github.com/consensys/accelerated-crypto-monorepo/maths/fft"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/collection"
	"github.com/consensys/accelerated-crypto-monorepo/utils/gnarkutil"
	"github.com/consensys/gnark/frontend"
)

/*
Same as `DeriveEvaluationPoint` but in a gnark circuit
*/
func GnarkDeriveEvaluationPoint(
	api frontend.API, h ifaces.Column, upstream string,
	cachedXs collection.Mapping[string, frontend.Variable],
	x frontend.Variable,
) (xRes []frontend.Variable) {

	if !h.IsComposite() {
		// Just return x and cache it if necessary
		newUpstream := appendNodeToUpstream(upstream, h)
		// Store in the cache if necessary
		if !cachedXs.Exists(newUpstream) {
			// Else register the result in the cache
			cachedXs.InsertNew(newUpstream, x)
		}
		return []frontend.Variable{x}
	}

	switch inner := h.(type) {
	case Shifted:
		newUpstream := appendNodeToUpstream(upstream, inner)
		var derivedX frontend.Variable
		// Early return if the result is cached
		if cachedXs.Exists(newUpstream) {
			derivedX = cachedXs.MustGet(newUpstream)
		} else {
			// If not, compute the shift on x and cache the result
			n := h.Size()
			omegaN := frontend.Variable(fft.GetOmega(n))
			omegaN = gnarkutil.Exp(api, omegaN, inner.Offset)
			derivedX = api.Mul(x, omegaN)
			cachedXs.InsertNew(newUpstream, derivedX)
		}
		return GnarkDeriveEvaluationPoint(api, inner.Parent, newUpstream, cachedXs, derivedX)

	case Repeated:
		newUpstream := appendNodeToUpstream(upstream, inner)
		var derivedX frontend.Variable
		// Early return if the result is cached
		if cachedXs.Exists(newUpstream) {
			derivedX = cachedXs.MustGet(newUpstream)
		} else {
			// If not, compute the shift on x and cache the result
			// Pass the exponentiation of the repeat
			derivedX = gnarkutil.Exp(api, x, inner.Nb)
			cachedXs.InsertNew(newUpstream, derivedX)
		}
		return GnarkDeriveEvaluationPoint(api, inner.Parent, newUpstream, cachedXs, derivedX)

	case Interleaved:
		// I(X) = sum_j<n/k : P1(X/rho_n^j) * (X^n - 1) / (X^k-rho_{n/k}^j)

		var res []frontend.Variable

		for j, parent := range inner.Parents {
			/*
				in order to avoid creating several queries for the same point
				we do not register any entry in the cache for j == 0. In this
				case we simply recurse the call, porting the same `upstream`
				and the same `x`.
			*/
			if j == 0 {
				fullyRecurseds := GnarkDeriveEvaluationPoint(
					api, parent, upstream, cachedXs, x,
				)
				res = append(res, fullyRecurseds...)
				continue
			}

			// Implictly, the id of each branch is associated to the
			// rational number i/n
			gcd := utils.GCD(j, len(inner.Parents))
			id := fmt.Sprintf("%v/%v", j/gcd, len(inner.Parents)/gcd)
			newUpstreamWithID := appendNodeToUpstreamWithID(upstream, inner, id)

			// Recompute the result or get it from the cache
			var derivedX frontend.Variable
			if cachedXs.Exists(newUpstreamWithID) {
				derivedX = cachedXs.MustGet(newUpstreamWithID)
			} else {
				n := h.Size()
				omegaN := frontend.Variable(fft.GetOmega(n))
				// Importantly, we need to pass
				derivedX = gnarkutil.Exp(api, omegaN, -j)
				derivedX = api.Mul(derivedX, x)
				// And cache them
				cachedXs.InsertNew(newUpstreamWithID, derivedX)
			}
			// Then recurse the call
			fullyRecurseds := GnarkDeriveEvaluationPoint(
				api, parent, newUpstreamWithID, cachedXs, derivedX,
			)
			res = append(res, fullyRecurseds...)

		}

		return res

	default:
		utils.Panic("unexpected type %v", reflect.TypeOf(inner))
	}
	panic("unreachable")
}

/*
Same as `VerifyYConsistency` but in a gnark circuit
*/
func GnarkVerifyYConsistency(
	api frontend.API, h ifaces.Column, upstream string,
	cachedXs collection.Mapping[string, frontend.Variable],
	finalYs collection.Mapping[string, frontend.Variable],
) (y frontend.Variable) {

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

	case Repeated:
		newUpstream := appendNodeToUpstream(upstream, inner)
		res := GnarkVerifyYConsistency(api, inner.Parent, newUpstream, cachedXs, finalYs)
		// No need to test the error, because we would return it alonside the
		// nil result anyway.
		return res

	case Interleaved:
		// I(X) = sum_j<n/k : P1(X/rho_n^j) * (X^n - 1) / (X^k-rho_{n/k}^j)
		// Get the xs, from the cache. We expect them to be here already
		xs := []frontend.Variable{}
		ys := []frontend.Variable{}

		// Get x from the cache
		for i, parent := range inner.Parents {
			/*
				Special-case : i == 0, take the upstream branch
				Regular-case : for the other branchs get them specifically from
				the cache.

				The disjunction helps avoiding creating several queries
				for the same x.
			*/
			newUpstream := upstream
			if i > 0 {
				// Implictly, the id of each branch is associated to the
				// rational number i/n.
				gcd := utils.GCD(i, len(inner.Parents))
				id := fmt.Sprintf("%v/%v", i/gcd, len(inner.Parents)/gcd)
				newUpstream = appendNodeToUpstreamWithID(upstream, inner, id)
			}

			x := cachedXs.MustGet(newUpstream)
			xs = append(xs, x)

			// Get the Y from downstream
			parentY := GnarkVerifyYConsistency(api, parent, newUpstream, cachedXs, finalYs)
			ys = append(ys, parentY)
		}

		// sanity-check
		if len(xs) != len(inner.Parents) {
			utils.Panic("There should be as many xs as there are parents")
		}

		res := inner.gnarkRecoverEvaluationFromParents(api, xs, ys)
		return res

	default:
		utils.Panic("unexpected type %v", reflect.TypeOf(inner))
	}
	panic("unreachable")
}
