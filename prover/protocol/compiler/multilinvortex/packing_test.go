package multilinvortex_test

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/multilinvortex"
	"github.com/stretchr/testify/require"
)

// baseTableToExt lifts a base-field table to fext for evaluation.
func baseTableToExt(t []field.Element) []fext.Element {
	out := make([]fext.Element, len(t))
	for i := range t {
		out[i].B0.A0 = t[i]
	}
	return out
}

// TestPackPolys_EquivalenceMath verifies the core packing equation:
//
//	P_i(ζ_i) = Q(b_i ‖ ζ_i)
//
// for randomly generated polynomials of mixed sizes and random evaluation
// points ζ_i. This validates the math of cross-size packing without involving
// the rest of the wizard.
func TestPackPolys_EquivalenceMath(t *testing.T) {
	cases := []struct {
		name  string
		sizes []int
	}{
		{"single_poly", []int{1 << 4}},
		{"two_equal", []int{1 << 5, 1 << 5}},
		{"two_different", []int{1 << 6, 1 << 4}},
		{"mixed_sizes", []int{1 << 7, 1 << 5, 1 << 5, 1 << 4, 1 << 3}},
		{"power_of_two_count", []int{1 << 6, 1 << 6, 1 << 6, 1 << 6}},
		{"non_pot_count_sum_fits", []int{1 << 4, 1 << 4, 1 << 4}},
		{"wide_spread", []int{1 << 10, 1 << 8, 1 << 6, 1 << 4, 1 << 4, 1 << 4}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rng := rand.New(rand.NewPCG(uint64(len(tc.sizes)), 0xC0DE))

			// Build K random polynomials of the given sizes.
			polys := make([][]field.Element, len(tc.sizes))
			for i, sz := range tc.sizes {
				polys[i] = make([]field.Element, sz)
				for j := range polys[i] {
					polys[i][j] = field.PseudoRand(rng)
				}
			}

			// Pack.
			p := multilinvortex.PackPolys(polys)
			require.NotNil(t, p)
			require.Equal(t, len(tc.sizes), len(p.Locators), "locator count")
			require.Equal(t, len(tc.sizes), len(p.Nv), "nv count")

			// Lift Q to fext for evaluation.
			QExt := baseTableToExt(p.Q)

			// For each poly, pick a random ζ_i and check the equivalence.
			for i, sz := range tc.sizes {
				nv := p.Nv[i]
				require.Equal(t, sz, 1<<nv, "poly %d size mismatch", i)

				// Random evaluation point in F_ext^nv.
				zeta := make([]fext.Element, nv)
				for j := range zeta {
					zeta[j] = randFext(rng)
				}

				// Direct evaluation: P_i(ζ_i).
				PiExt := baseTableToExt(polys[i])
				expected := evalMultilin(PiExt, zeta)

				// Packed evaluation: Q(b_i ‖ ζ_i).
				point := p.LocatorPoint(i, zeta)
				require.Equal(t, p.N, len(point), "extended point length")
				actual := evalMultilin(QExt, point)

				require.Truef(t, expected.Equal(&actual),
					"poly %d (nv=%d, locator=%d): P_i(ζ_i) = %v, Q(b_i‖ζ_i) = %v",
					i, nv, p.Locators[i], expected.String(), actual.String())
			}

			// Also verify locators are prefix-exclusive (no locator is a prefix
			// of another at the relevant lengths).
			for i := 0; i < len(tc.sizes); i++ {
				for j := i + 1; j < len(tc.sizes); j++ {
					require.Falsef(t, isPrefix(p.Locators[i], p.N-p.Nv[i], p.Locators[j], p.N-p.Nv[j]),
						"locator %d is a prefix of locator %d", i, j)
				}
			}
		})
	}
}

// isPrefix reports whether locator a (of length lenA bits) is a prefix of
// locator b (of length lenB bits) in MSB-first binary representation.
func isPrefix(a, lenA, b, lenB int) bool {
	if lenA > lenB {
		return false
	}
	if lenA == 0 {
		return false // empty locator can't be a "prefix" in this packing scheme
	}
	// Take top lenA bits of b and compare to a.
	topB := b >> (lenB - lenA)
	return topB == a
}

// randFext returns a random fext.Element drawn from a PCG.
func randFext(rng *rand.Rand) fext.Element {
	var x fext.Element
	x.B0.A0 = field.PseudoRand(rng)
	x.B0.A1 = field.PseudoRand(rng)
	x.B1.A0 = field.PseudoRand(rng)
	x.B1.A1 = field.PseudoRand(rng)
	return x
}
