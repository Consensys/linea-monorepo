// Package multilinvortex — cross-size packing primitives.
//
// This file implements the prefix-exclusive locator packing scheme described
// in the protocol spec: given polynomials P_0, ..., P_{K-1} of (possibly
// distinct) power-of-2 sizes 2^{n_i}, pack them into a single polynomial Q
// of size 2^N such that for each i:
//
//	P_i(z_0, ..., z_{n_i - 1}) = Q(b_{i,0}, ..., b_{i,N-n_i-1}, z_0, ..., z_{n_i-1})
//
// where b_i ∈ {0,1}^{N-n_i} is a prefix-exclusive binary locator. Prefix-
// exclusiveness ensures the P_i occupy disjoint sub-spaces of Q, so they can
// be recovered without overlap. The locators form the branches of a binary
// tree whose leaves are the slots reserved for each P_i.
//
// Equivalent claim conversion: the verifier's ML evaluation claim
// P_i(ζ_i) = y_i is the same as Q(b_i ‖ ζ_i) = y_i.
//
// This is the multi-size generalisation of the per-size-group packing used
// by CompileRoundPacked. With this primitive, all K polynomials across all
// size groups can in principle be packed into ONE Vortex commitment.

package multilinvortex

import (
	"fmt"
	"math/bits"
	"sort"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
)

// PackedPolys is the result of packing K polynomials of distinct power-of-2
// sizes into a single polynomial Q of size 2^N.
type PackedPolys struct {
	// Q is the packed polynomial as a flat 2^N base-field vector.
	// Positions not covered by any P_i are zero.
	Q []field.Element
	// N is the dimension of Q (i.e. |Q| = 2^N).
	N int
	// Locators[i] is the integer encoding of poly i's prefix-exclusive
	// locator. The binary representation, padded on the left to (N - n_i)
	// bits, gives b_i = (b_{i,0}, ..., b_{i,N-n_i-1}) in MSB-first order.
	Locators []int
	// Nv[i] is the number of variables of poly i (i.e. |P_i| = 2^Nv[i]).
	Nv []int
}

// PackPolys packs the given polynomials into a single polynomial Q via
// prefix-exclusive locator coding using a buddy allocator. Polys may have
// different sizes; each size must be a power of 2.
//
// Allocation policy: sort by size descending, greedy first-fit at the
// smallest available level. This is optimal for the buddy scheme — the total
// space used equals the smallest power-of-2 ≥ Σ 2^{n_i}.
func PackPolys(polys [][]field.Element) *PackedPolys {
	K := len(polys)
	if K == 0 {
		return &PackedPolys{}
	}

	var total int
	nv := make([]int, K)
	for k, p := range polys {
		if len(p) == 0 {
			continue
		}
		if len(p)&(len(p)-1) != 0 {
			panic(fmt.Sprintf("PackPolys: poly %d size %d is not a power of 2", k, len(p)))
		}
		nv[k] = bits.TrailingZeros(uint(len(p)))
		total += len(p)
	}
	if total == 0 {
		return &PackedPolys{Nv: nv}
	}

	N := bits.Len(uint(total - 1))
	if N == 0 {
		N = 1
	}
	Q := make([]field.Element, 1<<N)

	// Buddy allocator: free[level] is the stack of free offsets at size 2^level.
	free := make([][]int, N+1)
	free[N] = []int{0}

	var allocate func(level int) int
	allocate = func(level int) int {
		if level > N {
			return -1
		}
		if n := len(free[level]); n > 0 {
			off := free[level][n-1]
			free[level] = free[level][:n-1]
			return off
		}
		parent := allocate(level + 1)
		if parent == -1 {
			return -1
		}
		// Split: use first half, free second half.
		free[level] = append(free[level], parent+(1<<level))
		return parent
	}

	// Sort indices by size descending to bin-pack large polys first.
	order := make([]int, K)
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool {
		return len(polys[order[i]]) > len(polys[order[j]])
	})

	locators := make([]int, K)
	for _, k := range order {
		size := len(polys[k])
		if size == 0 {
			continue
		}
		level := nv[k]
		off := allocate(level)
		if off == -1 {
			panic(fmt.Sprintf("PackPolys: no space for poly %d (size %d) in Q (size %d)", k, size, 1<<N))
		}
		copy(Q[off:off+size], polys[k])
		locators[k] = off >> level
	}

	return &PackedPolys{Q: Q, N: N, Locators: locators, Nv: nv}
}

// LocatorPoint returns the (N+|rest|)-dimensional locator-extended evaluation
// point for poly i: b_i ‖ rest, where b_i is the (N - nv_i)-bit big-endian
// binary encoding of the locator integer.
//
// This is the multi-size analogue of locatorPoint (which only handles the
// equal-size case).
func (p *PackedPolys) LocatorPoint(i int, rest []fext.Element) []fext.Element {
	L := p.N - p.Nv[i]
	if L < 0 {
		panic(fmt.Sprintf("LocatorPoint: N=%d < nv[%d]=%d", p.N, i, p.Nv[i]))
	}
	pt := make([]fext.Element, L+len(rest))
	loc := p.Locators[i]
	for j := 0; j < L; j++ {
		if (loc>>(L-1-j))&1 == 1 {
			pt[j].SetOne()
		}
	}
	copy(pt[L:], rest)
	return pt
}
