package gkrverifier

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	extensionsHash "github.com/consensys/gnark-crypto/field/koalabear/extensions/hash"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/gkr"
	gkrkoalabear "github.com/consensys/gnark/gkr/koalabear"
	gkrposeidon2 "github.com/consensys/gnark/gkr/permutation/gkr-poseidon2/koalabear"
)

const verifyDriver = "../../src/test/zkc/gkr/verify.zkc"

// writeAllTables writes the gate, Lagrange and schedule tables of the circuit to files in a
// temporary directory, and returns their paths.
func writeAllTables(t *testing.T, circuit gkr.SerializableCircuit, schedule constraint.GkrProvingSchedule) []string {
	t.Helper()
	dir := t.TempDir()
	writers := map[string]func(f *os.File) error{
		"gates.zkc":    func(f *os.File) error { _, err := WriteGateTables(f, circuit); return err },
		"lagrange.zkc": func(f *os.File) error { return WriteLagrangeTables(f, circuit, schedule) },
		"schedule.zkc": func(f *os.File) error { return WriteScheduleTables(f, circuit, schedule) },
	}
	var paths []string
	for _, name := range []string{"gates.zkc", "lagrange.zkc", "schedule.zkc"} {
		path := filepath.Join(dir, name)
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := writers[name](f); err != nil {
			t.Fatal(err)
		}
		f.Close()
		paths = append(paths, path)
	}
	return paths
}

func newHasher() *extensionsHash.BaseFieldHashAdapterE6 {
	return extensionsHash.NewBaseFieldHashAdapterE6(poseidon2.NewMerkleDamgardHasher())
}

// proveCompressor returns a proof of the compressor on 2ⁿ random instances, made with the hasher, with
// the claims that proving returns.
func proveCompressor(t *testing.T, compressor *gkrposeidon2.Compressor, n int, hasher *extensionsHash.BaseFieldHashAdapterE6, rng *rand.Rand) (gkrkoalabear.Proof, gkrkoalabear.Claims) {
	t.Helper()
	half := len(compressor.Left)
	columns := func() [][]koalabear.Element {
		res := make([][]koalabear.Element, half)
		for i := range res {
			res[i] = make([]koalabear.Element, 1<<n)
			for j := range res[i] {
				res[i][j].SetUint64(uint64(rng.Uint32()))
			}
		}
		return res
	}
	proof, claims, err := gkrkoalabear.Prove(compressor.Circuit, compressor.Schedule, compressor.Assignment(columns(), columns()), hasher)
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

// verifyInstance is a proof, with its seed and its claims, as the ZkC verifier takes them.
type verifyInstance struct {
	n      int
	seed   []extensions.E6
	proof  []extensions.E6
	claims []extensions.E6 // for each claim, its point, then its value
}

func (v *verifyInstance) inputs() map[string]zkcInput {
	words := func(es []extensions.E6) zkcInput {
		var res zkcInput
		for i := range es {
			res = append(res, e6Words(&es[i])...)
		}
		return res
	}
	return map[string]zkcInput{
		"gkr_n":       {uint32(v.n)},
		"gkr_nb_seed": {uint32(len(v.seed))},
		"gkr_seed":    words(v.seed),
		"gkr_proof":   words(v.proof),
		"gkr_claims":  words(v.claims),
	}
}

func TestVerify(t *testing.T) {
	compressor, err := gkrposeidon2.NewCompressor(poseidon2.GetDefaultParameters())
	if err != nil {
		t.Fatal(err)
	}
	circuit, schedule := compressor.Circuit, compressor.Schedule
	tables, err := newScheduleTables(circuit, schedule)
	if err != nil {
		t.Fatal(err)
	}
	if len(tables.eqs) != 256 {
		t.Errorf("%d eq entries, expected 256", len(tables.eqs))
	}
	rng := rand.New(rand.NewPCG(13, 14))
	paths := writeAllTables(t, circuit, schedule)
	run := func(inputs map[string]zkcInput) bool { return runZkc(t, inputs, verifyDriver, paths...) }

	newInstance := func(n int) *verifyInstance {
		v := &verifyInstance{n: n, seed: make([]extensions.E6, 3)}
		for i := range v.seed {
			v.seed[i] = randomE6(rng)
		}
		seeded := func() *extensionsHash.BaseFieldHashAdapterE6 {
			h := newHasher()
			h.WriteField(v.seed...)
			return h
		}
		proof, claims := proveCompressor(t, compressor, n, seeded(), rng)
		v.proof = flatten(proof)
		if _, err := gkrkoalabear.Verify(circuit, schedule, n, proof, seeded()); err != nil {
			t.Fatal(err)
		}
		used := make(map[int]int)
		for _, r := range tables.returned {
			claim := claims[r.wire][used[r.wire]]
			used[r.wire]++
			v.claims = append(v.claims, claim.EvaluationPoint...)
			v.claims = append(v.claims, claim.Evaluation)
		}
		return v
	}

	sizes := []int{1, 2, 5}
	if !testing.Short() {
		sizes = append(sizes, 16)
	}
	for _, n := range sizes {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			v := newInstance(n)
			if !run(v.inputs()) {
				t.Fatal("correct proof rejected")
			}
			if n == 16 {
				return
			}

			// reject runs the verifier on inputs with the word number word of the input array
			// changed to value, and reports whether it was rejected.
			rejects := func(name, array string, word int, change func(uint32) uint32) {
				inputs := v.inputs()
				inputs[array][word] = change(inputs[array][word])
				if run(inputs) {
					t.Errorf("%s: accepted", name)
				}
			}
			addOne := func(w uint32) uint32 { return (w + 1) % modulusU32 }

			firstOf := func(kind int) *levelTable {
				for i := range tables.levels {
					if tables.levels[i].kind == kind {
						return &tables.levels[i]
					}
				}
				t.Fatalf("no level of kind %d", kind)
				return nil
			}
			proofWord := func(p position) int { return 6 * p.at(n) }

			rejects("a round polynomial value at a sumcheck level", "gkr_proof", proofWord(firstOf(levelSumcheck).start), addOne)
			rejects("a round polynomial value at a single-source level", "gkr_proof", proofWord(firstOf(levelSingleSource).start)+6*n, addOne)
			rejects("a final evaluation at a skip level", "gkr_proof", proofWord(firstOf(levelSkip).finals), addOne)
			rejects("an output evaluation", "gkr_proof", proofWord(tables.levels[len(schedule)].start), addOne)
			rejects("a final evaluation at level 0", "gkr_proof", proofWord(tables.levels[0].finals), addOne)

			// Changing a final evaluation of level 0 and the claims that return it leaves only the
			// level's own final check to reject.
			consistent := v.inputs()
			element := tables.levels[0].finals.at(n)
			consistent["gkr_proof"][6*element] = addOne(consistent["gkr_proof"][6*element])
			for i, r := range tables.returned {
				if r.value.at(n) == element {
					word := 6 * (i*(n+1) + n)
					consistent["gkr_claims"][word] = addOne(consistent["gkr_claims"][word])
				}
			}
			if run(consistent) {
				t.Error("a final evaluation at level 0, with the claims changed along: accepted")
			}

			rejects("the seed", "gkr_seed", 7, addOne)
			rejects("a coordinate of a claimed point", "gkr_claims", 3, addOne)
			rejects("a claimed value", "gkr_claims", 6*n, addOne)
			rejects("a non-canonical proof word", "gkr_proof", 0, func(w uint32) uint32 { return w + modulusU32 })
			rejects("a non-canonical seed word", "gkr_seed", 0, func(w uint32) uint32 { return w + modulusU32 })
			rejects("a non-canonical claim word", "gkr_claims", 0, func(w uint32) uint32 { return w + modulusU32 })
			rejects("n - 1", "gkr_n", 0, func(w uint32) uint32 { return w - 1 })

			// With n + 1 the verifier reads more of the proof and the claims than they hold, and zkc
			// panics on a read of an input past its end, so they are padded with zeros.
			padded := v.inputs()
			padded["gkr_n"][0]++
			for array, length := range map[string]int{
				"gkr_proof":  6 * (tables.proof.a*(n+1) + tables.proof.b),
				"gkr_claims": 6 * len(tables.returned) * (n + 2),
			} {
				padded[array] = append(padded[array], make(zkcInput, length-len(padded[array]))...)
			}
			if run(padded) {
				t.Error("n + 1: accepted")
			}
		})
	}
}
