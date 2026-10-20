package gkrverifier

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	extensionsHash "github.com/consensys/gnark-crypto/field/koalabear/extensions/hash"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark/constraint"
	gkrkoalabear "github.com/consensys/gnark/gkr/koalabear"
	gkrposeidon2 "github.com/consensys/gnark/gkr/permutation/gkr-poseidon2/koalabear"
)

func newHasher() *extensionsHash.BaseFieldHashAdapterE6 {
	return extensionsHash.NewBaseFieldHashAdapterE6(poseidon2.NewMerkleDamgardHasher())
}

// proveCompressor returns a proof of the compressor on 2ⁿ random instances, with the claims that
// proving returns.
func proveCompressor(t *testing.T, compressor *gkrposeidon2.Compressor, n int) (gkrkoalabear.Proof, gkrkoalabear.Claims) {
	t.Helper()
	half := len(compressor.Left)
	columns := func() [][]koalabear.Element {
		res := make([][]koalabear.Element, half)
		for i := range res {
			res[i] = make([]koalabear.Element, 1<<n)
			koalabear.Vector(res[i]).MustSetRandom()
		}
		return res
	}
	proof, claims, err := gkrkoalabear.Prove(compressor.Circuit, compressor.Schedule, compressor.Assignment(columns(), columns()), newHasher())
	if err != nil {
		t.Fatal(err)
	}
	return proof, claims
}

func flatten(proof gkrkoalabear.Proof) []extensions.E6 {
	var res []extensions.E6
	for _, e := range proof.Flatten() {
		res = append(res, *e)
	}
	return res
}

func (p position) at(n int) int { return p.a*n + p.b }

func TestTables(t *testing.T) {
	compressor, err := gkrposeidon2.NewCompressor(poseidon2.GetDefaultParameters())
	if err != nil {
		t.Fatal(err)
	}
	circuit, schedule := compressor.Circuit, compressor.Schedule
	tables, err := newScheduleTables(circuit, schedule)
	if err != nil {
		t.Fatal(err)
	}
	nbLevels := len(schedule)

	t.Run("sizes", func(t *testing.T) {
		nbSumcheckClaims, nbEqs, nbSumcheckLevels := 0, 0, 0
		for levelI, level := range schedule {
			if _, ok := level.(*constraint.GkrSumcheckLevel); !ok {
				if tables.levels[levelI].nbEq != 0 {
					t.Errorf("level %d has eq entries, and is not a sumcheck level", levelI)
				}
				continue
			}
			nbSumcheckLevels++
			nbSumcheckClaims += level.NbClaims()
			for _, group := range level.ClaimGroups() {
				nbEqs += len(group.ClaimSources)
			}
		}
		t.Logf("%d sumcheck levels, %d claims, %d eq entries", nbSumcheckLevels, nbSumcheckClaims, len(tables.eqs))
		if nbSumcheckClaims != 614 {
			t.Errorf("%d claims at sumcheck levels, expected 614", nbSumcheckClaims)
		}
		if len(tables.eqs) != 256 || nbEqs != 256 {
			t.Errorf("%d eq entries, expected 256, and %d claim sources", len(tables.eqs), nbEqs)
		}
		t.Logf("%d points, proof of %d·n + %d elements, n at most %d", tables.nbPoints, tables.proof.a, tables.proof.b, tables.nMax)
	})

	for _, n := range []int{1, 2, 3} {
		t.Run(fmt.Sprintf("proof n=%d", n), func(t *testing.T) {
			proof, claims := proveCompressor(t, compressor, n)
			flat := flatten(proof)
			if want := tables.proof.at(n); len(flat) != want {
				t.Fatalf("the proof has %d elements, expected %d", len(flat), want)
			}

			verify := func(flat []extensions.E6) (gkrkoalabear.Claims, error) {
				proof, err := gkrkoalabear.DeserializeProof(circuit, schedule, flat)
				if err != nil {
					t.Fatal(err)
				}
				return gkrkoalabear.Verify(circuit, schedule, n, proof, newHasher())
			}
			verifierClaims, err := verify(flat)
			if err != nil {
				t.Fatal(err)
			}

			// A change to the first element of a level makes the verification fail at that level, and
			// a change to the last element of the level before makes it fail at that one.
			failsAt := func(p int) int {
				changed := slices.Clone(flat)
				var one extensions.E6
				one.SetOne()
				changed[p].Add(&changed[p], &one)
				_, err := verify(changed)
				if err == nil {
					t.Fatalf("a change to element %d is accepted", p)
				}
				var level int
				if _, scanErr := fmt.Sscanf(err.Error(), "level %d:", &level); scanErr != nil {
					t.Fatalf("a change to element %d fails with %q", p, err)
				}
				return level
			}
			for levelI := range nbLevels {
				start := tables.levels[levelI].start.at(n)
				if got := failsAt(start); got != levelI {
					t.Errorf("level %d: a change to its first element, %d, fails at level %d", levelI, start, got)
				}
				if levelI != 0 {
					if got := failsAt(start - 1); got != levelI-1 {
						t.Errorf("level %d: a change to the last element of the level before, %d, fails at level %d", levelI, start-1, got)
					}
				}
			}
			if got, want := tables.levels[nbLevels].start.at(n), len(flat)-len(circuit.Outputs()); got != want {
				t.Errorf("the output level starts at %d, expected %d", got, want)
			}

			// At skip levels the gate of each wire on the final evaluations is the claimed value.
			for levelI, l := range tables.levels[:nbLevels] {
				if l.kind != levelSkip {
					continue
				}
				claimRow := l.firstClaim
				for _, w := range tables.wires[l.firstWire : l.firstWire+l.nbWires] {
					for k := range w.nbSources {
						inputs := make([]extensions.E6, w.nbIn)
						for i, u := range tables.inputs[w.firstInput : w.firstInput+w.nbIn] {
							inputs[i] = flat[l.finals.at(n)+u*l.nbOut+k]
						}
						gate := evaluateGate(tables.gates.gates[w.gate], inputs)
						if claimed := flat[tables.claims[claimRow].value.at(n)]; !claimed.Equal(&gate) {
							t.Errorf("level %d: claim row %d is not the gate's value", levelI, claimRow)
						}
						claimRow++
					}
				}
			}

			// The returned claims
			if len(tables.returned) != claimsCount(verifierClaims) || claimsCount(claims) != claimsCount(verifierClaims) {
				t.Fatalf("%d returned rows, %d claims verified, %d proved", len(tables.returned), claimsCount(verifierClaims), claimsCount(claims))
			}
			used := make(map[int]int) // wire → number of its claims so far
			points := make(map[int][]extensions.E6)
			for i, r := range tables.returned {
				claim := verifierClaims[r.wire][used[r.wire]]
				used[r.wire]++
				if !claim.Evaluation.Equal(&flat[r.value.at(n)]) {
					t.Errorf("returned claim %d (wire %d): the value is not at its position", i, r.wire)
				}
				if point, seen := points[r.point]; !seen {
					points[r.point] = claim.EvaluationPoint
				} else if !slices.EqualFunc(point, claim.EvaluationPoint, func(a, b extensions.E6) bool { return a.Equal(&b) }) {
					t.Errorf("returned claim %d (wire %d): point %d differs from an earlier claim's", i, r.wire, r.point)
				}
			}
			for p, point := range points {
				for q, other := range points {
					if p < q && slices.EqualFunc(point, other, func(a, b extensions.E6) bool { return a.Equal(&b) }) {
						t.Errorf("points %d and %d are equal", p, q)
					}
				}
			}
		})
	}

	t.Run("generated tables", func(t *testing.T) {
		var sb strings.Builder
		if err := WriteScheduleTables(&sb, circuit, schedule); err != nil {
			t.Fatal(err)
		}
		t.Logf("%d bytes", sb.Len())
	})
}

func claimsCount(c gkrkoalabear.Claims) int {
	n := 0
	for _, claims := range c {
		n += len(claims)
	}
	return n
}
