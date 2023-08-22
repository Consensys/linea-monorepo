package column

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/consensys/accelerated-crypto-monorepo/maths/fft"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/collection"
)

/*
	Various functions to represent handles compositions
*/

/*
Given a query at point "x" on the current handle, returns the list
of "x" values we should query the "leaves" of the current handle
to get an equivalent polynomial query only on "Non-Composites" handles.
The result is stored in the map `finalXs` which will only contain
the "leaves" derived Xs as opposed to the `cachedXs` which computes
all intermediate values.

  - upstreamBranchRepr string : repr of the upstream-branch
  - cachedXs : map reprs of full-branches to already derived
    evaluation points. Essentially a cache for later uses.
  - x is the 'x' value of the univariate query to evaluate
*/
func DeriveEvaluationPoint(
	h ifaces.Column,
	upstream string,
	cachedXs collection.Mapping[string, field.Element],
	x field.Element,
) (xRes []field.Element) {

	if !h.IsComposite() {
		// Just return x and cache it if necessary
		newUpstream := appendNodeToUpstream(upstream, h)
		// Store in the cache if necessary
		if !cachedXs.Exists(newUpstream) {
			// Else register the result in the cache
			cachedXs.InsertNew(newUpstream, x)
		}
		return []field.Element{x}
	}

	switch inner := h.(type) {
	case Shifted:
		newUpstream := appendNodeToUpstream(upstream, inner)
		var derivedX field.Element
		// Early return if the result is cached
		if cachedXs.Exists(newUpstream) {
			derivedX = cachedXs.MustGet(newUpstream)
		} else {
			// If not, compute the shift on x and cache the result
			n := h.Size()
			omegaN := fft.GetOmega(n)
			omegaN.Exp(omegaN, big.NewInt(int64(inner.Offset)))
			derivedX.Mul(&x, &omegaN)
			cachedXs.InsertNew(newUpstream, derivedX)
		}
		return DeriveEvaluationPoint(inner.Parent, newUpstream, cachedXs, derivedX)

	case Repeated:
		newUpstream := appendNodeToUpstream(upstream, inner)
		var derivedX field.Element
		// Early return if the result is cached
		if cachedXs.Exists(newUpstream) {
			derivedX = cachedXs.MustGet(newUpstream)
		} else {
			// If not, compute the shift on x and cache the result
			// Pass the exponentiation of the repeat
			derivedX.Exp(x, big.NewInt(int64(inner.Nb)))
			cachedXs.InsertNew(newUpstream, derivedX)
		}
		return DeriveEvaluationPoint(inner.Parent, newUpstream, cachedXs, derivedX)

	case Interleaved:
		// I(X) = sum_j<n/k : P1(X/rho_n^j) * (X^n - 1) / (X^k-rho_{n/k}^j)

		var res []field.Element

		for j, parent := range inner.Parents {
			/*
				in order to avoid creating several queries for the same point
				we do not register any entry in the cache for j == 0. In this
				case we simply recurse the call, porting the same `upstream`
				and the same `x`.
			*/
			if j == 0 {
				fullyRecurseds := DeriveEvaluationPoint(
					parent, upstream, cachedXs, x,
				)
				res = append(res, fullyRecurseds...)
				continue
			}

			// implictly, the id of each branch is associated to the
			// rational number i/n
			gcd := utils.GCD(j, len(inner.Parents))
			id := fmt.Sprintf("%v/%v", j/gcd, len(inner.Parents)/gcd)
			newUpstreamWithID := appendNodeToUpstreamWithID(upstream, inner, id)

			// Recompute the result or get it from the cache
			var derivedX field.Element
			if cachedXs.Exists(newUpstreamWithID) {
				derivedX = cachedXs.MustGet(newUpstreamWithID)
			} else {
				n := h.Size()
				omegaN := fft.GetOmega(n)
				// Importantly, we need to pass
				derivedX.Exp(omegaN, big.NewInt(int64(-j)))
				derivedX.Mul(&derivedX, &x)
				// And cache them
				cachedXs.InsertNew(newUpstreamWithID, derivedX)
			}
			// Then recurse the call
			fullyRecurseds := DeriveEvaluationPoint(
				parent, newUpstreamWithID, cachedXs, derivedX,
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
Verifies that the claimed values for y
  - upstream is a repr of the branch leading to the current node
  - cachedXs is the set of derived Xs computed by `DeriveEvaluationPoint`
  - finalYs is the map of the alleged evaluation for which we check the
    consistency.
*/
func VerifyYConsistency(
	h ifaces.Column, upstream string,
	cachedXs collection.Mapping[string, field.Element],
	finalYs collection.Mapping[string, field.Element],
) (y field.Element) {

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
		res := VerifyYConsistency(inner.Parent, newUpstream, cachedXs, finalYs)
		// No need to test the error, because we would return it alonside the
		// nil result anyway.
		return res

	case Repeated:
		newUpstream := appendNodeToUpstream(upstream, inner)
		res := VerifyYConsistency(inner.Parent, newUpstream, cachedXs, finalYs)
		// No need to test the error, because we would return it alonside the
		// nil result anyway.
		return res

	case Interleaved:
		// I(X) = sum_j<n/k : P1(X/rho_n^j) * (X^n - 1) / (X^k-rho_{n/k}^j)
		// Get the xs, from the cache. We expect them to be here already
		xs := []field.Element{}
		ys := []field.Element{}

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
				// rational number i/n
				gcd := utils.GCD(i, len(inner.Parents))
				id := fmt.Sprintf("%v/%v", i/gcd, len(inner.Parents)/gcd)
				newUpstream = appendNodeToUpstreamWithID(upstream, inner, id)
			}

			x := cachedXs.MustGet(newUpstream)
			xs = append(xs, x)

			// Get the Ys from downstream
			parentY := VerifyYConsistency(parent, newUpstream, cachedXs, finalYs)
			ys = append(ys, parentY)
		}

		// sanity-check
		if len(xs) != len(inner.Parents) {
			utils.Panic("There should be as many xs as there are parents")
		}

		res := inner.recoverEvaluationFromParents(xs, ys)
		return res

	default:
		utils.Panic("unexpected type %v", reflect.TypeOf(inner))
	}
	panic("unreachable")
}

/*
Returns all subbranches starting from the current node
(including the current node). It counts as a list of unique
identifier for derivation path.
*/
func AllDownStreamBranches(node ifaces.Column) []string {

	if !node.IsComposite() {
		return []string{getNodeRepr(node)}
	}

	switch inner := node.(type) {
	case Repeated:
		downStreams := AllDownStreamBranches(inner.Parent)
		return prependNodeToDownstream(inner, downStreams)
	case Shifted:
		downStreams := AllDownStreamBranches(inner.Parent)
		return prependNodeToDownstream(inner, downStreams)
	case Interleaved:
		res := []string{}
		for i, p := range inner.Parents {
			// In the case i == 0, the evaluation point is the same as if the
			// downstream branch was not part of an interleaving. For this reason
			// we skip it here.
			downStreams := AllDownStreamBranches(p)
			if i > 0 {
				// Implictly, the id of each branch is associated to the
				// rational number i/n
				gcd := utils.GCD(i, len(inner.Parents))
				id := fmt.Sprintf("%v/%v", i/gcd, len(inner.Parents)/gcd)
				downStreams = prependNodeToDownstreamWithID(inner, id, downStreams)
			}
			res = append(res, downStreams...)
		}
		return res
	default:
		panic("unreachable")
	}
}

func appendNodeToUpstream(upstream string, node ifaces.Column) string {
	// This happens when the function is called on the root node
	// In this case, no need to prefix with a "_"
	if len(upstream) == 0 {
		return getNodeRepr(node)
	}
	return fmt.Sprintf("%v_%v", upstream, getNodeRepr(node))
}

func appendNodeToUpstreamWithID(upstream string, node Interleaved, id string) string {

	// This happens when the function is called on the root node
	// In this case, no need to prefix with a "_"
	if len(upstream) == 0 {
		return fmt.Sprintf("%v_%v", getNodeRepr(node), id)
	}
	return fmt.Sprintf("%v_%v_%v", upstream, getNodeRepr(node), id)
}

func prependNodeToDownstream(node ifaces.Column, downstream []string) []string {
	res := []string{}
	nodeRepr := getNodeRepr(node)
	for _, d := range downstream {
		newbranch := fmt.Sprintf("%v_%v", nodeRepr, d)
		res = append(res, newbranch)
	}
	return res
}

func prependNodeToDownstreamWithID(node ifaces.Column, id string, downstream []string) []string {
	res := []string{}
	nodeRepr := getNodeRepr(node)
	for _, d := range downstream {
		newbranch := fmt.Sprintf("%v_%v_%v", nodeRepr, id, d)
		res = append(res, newbranch)
	}
	return res
}

func DerivedYRepr(upstream string, currNode ifaces.Column) string {
	ifaces.AssertNotComposite(currNode)
	// sanity-check
	if !strings.HasSuffix(upstream, NONCOMPOSITE) {
		utils.Panic("upstream was not suffixed by `NONCOMPOSITE`")
	}
	return fmt.Sprintf("%v_%v",
		upstream,
		currNode.GetColID(),
	)
}

func getNodeRepr(node ifaces.Column) string {

	if !node.IsComposite() {
		return NONCOMPOSITE
	}

	switch inner := node.(type) {
	case Repeated:
		return fmt.Sprintf("%v_%v", REPEAT, inner.Nb)
	case Shifted:
		// In a shifted, the size matters because this tells us which root of
		// unity to use to multiply `x`.
		return fmt.Sprintf("%v_%v_%v", SHIFT, utils.PositiveMod(inner.Offset, node.Size()), node.Size())
	case Interleaved:
		// In a shifted, the  of the parents matters. size matters because this
		// tells us which root of unity to use to multiply `x`.
		return fmt.Sprintf("%v_%v", INTERLEAVED, inner.Parents[0].Size())
	case ifaces.Column:
		// Common unexpectable case
		panic("don't pass a ifaces.Column there, pass the inner directly")
	default:
		utils.Panic("unexpected type %v\n", reflect.TypeOf(inner).String())
	}
	panic("unreachable")
}
