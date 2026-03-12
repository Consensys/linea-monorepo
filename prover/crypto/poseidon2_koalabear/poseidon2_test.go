package poseidon2_koalabear

import (
	"fmt"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	_ "github.com/consensys/gnark/std/hash/all"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"
)

// wideCommitBuilder wraps a U32 builder to implement frontend.WideCommitter,
// which is required for multi-instance GKR on small fields.
// Mirrors gnark/internal/widecommitter.
type wideCommitBuilder struct {
	frontend.Builder[constraint.U32]
}

func wideCommitWrapper(newBuilder frontend.NewBuilderU32) frontend.NewBuilderU32 {
	return func(field *big.Int, config frontend.CompileConfig) (frontend.Builder[constraint.U32], error) {
		b, err := newBuilder(field, config)
		if err != nil {
			return nil, err
		}
		return &wideCommitBuilder{b}, nil
	}
}

func (w *wideCommitBuilder) WideCommit(width int, toCommit ...frontend.Variable) ([]frontend.Variable, error) {
	return w.NewHint(wideCommitHint, width, toCommit...)
}

func wideCommitHint(m *big.Int, inputs []*big.Int, outputs []*big.Int) error {
	nb := (m.BitLen() + 7) / 8
	buf := make([]byte, nb)
	hasher := sha3.NewCShake128(nil, []byte("gnark test engine"))
	for _, in := range inputs {
		bs := in.FillBytes(buf)
		hasher.Write(bs)
	}
	for i := range outputs {
		hasher.Read(buf)
		outputs[i].SetBytes(buf)
		outputs[i].Mod(outputs[i], m)
	}
	return nil
}

func init() {
	solver.RegisterHint(wideCommitHint)
}

//---------------------------------------
// native variables

type GnarkMDHasherCircuit struct {
	Inputs []frontend.Variable
	Ouput  GnarkOctuplet
}

func (ghc *GnarkMDHasherCircuit) Define(api frontend.API) error {

	h, err := NewGnarkMDHasher(api)
	if err != nil {
		return err
	}

	// write elmts
	h.Write(ghc.Inputs...)

	// sum
	res := h.Sum()

	// check the result
	for i := 0; i < 8; i++ {
		api.AssertIsEqual(ghc.Ouput[i], res[i])
	}

	return nil
}

func getGnarkMDHasherCircuitWitness(nbElmts int) (*GnarkMDHasherCircuit, *GnarkMDHasherCircuit) {

	// values to hash
	vals := make([]field.Element, nbElmts)
	for i := 0; i < nbElmts; i++ {
		vals[i].SetRandom()
	}

	// sum
	phasher := NewMDHasher()
	phasher.WriteElements(vals...)
	res := phasher.SumElement()

	// create witness and circuit
	var circuit, witness GnarkMDHasherCircuit
	circuit.Inputs = make([]frontend.Variable, nbElmts)
	witness.Inputs = make([]frontend.Variable, nbElmts)
	for i := 0; i < nbElmts; i++ {
		witness.Inputs[i] = vals[i].String()
	}
	for i := 0; i < 8; i++ {
		witness.Ouput[i] = res[i].String()
	}

	return &circuit, &witness
}

// TestCircuitCompile verifies the GKR-backed Poseidon2 circuit compiles.
func TestCircuitCompile(t *testing.T) {
	require.NoError(t, RegisterGates())

	for _, nbElmts := range []int{8, 16, 444} {
		t.Run(fmt.Sprintf("Size_%d", nbElmts), func(t *testing.T) {
			circuit, _ := getGnarkMDHasherCircuitWitness(nbElmts)

			ccs, err := frontend.CompileU32(koalabear.Modulus(), wideCommitWrapper(scs.NewBuilder), circuit)
			require.NoError(t, err)
			fmt.Printf("GKR Poseidon2 (%d elems, %d instances): %d constraints\n",
				nbElmts, (nbElmts+7)/8, ccs.GetNbConstraints())
		})
	}
}

// TestCircuitSolve tests the full compile + solve cycle.
func TestGKRCircuitSolve(t *testing.T) {
	require.NoError(t, RegisterGates())

	for _, nbElmts := range []int{8, 16, 444, 800, 160000} {
		t.Run(fmt.Sprintf("Size_%d", nbElmts), func(t *testing.T) {
			circuit, witness := getGnarkMDHasherCircuitWitness(nbElmts)

			t0 := time.Now()
			ccs, err := frontend.CompileU32(koalabear.Modulus(), wideCommitWrapper(scs.NewBuilder), circuit)
			require.NoError(t, err)
			compileElapsed := time.Since(t0)

			fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
			require.NoError(t, err)
			t1 := time.Now()
			err = ccs.IsSolved(fullWitness)
			require.NoError(t, err)
			solveElapsed := time.Since(t1)
			fmt.Printf("[gkr]    size=%d instances=%d constraints=%d compile=%s solve=%s\n",
				nbElmts, (nbElmts+7)/8, ccs.GetNbConstraints(), compileElapsed, solveElapsed)
		})
	}
}

// nativeMDHasherCircuit tests the plain (non-GKR) Poseidon2 gnark circuit.
// Constructs GnarkMDHasher directly (gkrCompressor=nil) so Sum() calls
// CompressPoseidon2 from poseidon2_circuit.go with no GKR batching.
type nativeMDHasherCircuit struct {
	Inputs []frontend.Variable
	Output GnarkOctuplet
}

func (c *nativeMDHasherCircuit) Define(api frontend.API) error {
	// Construct directly so gkrCompressor=nil → Sum() uses CompressPoseidon2.
	// Must initialize state to 0 (nil frontend.Variable panics in Add).
	var state GnarkOctuplet
	for i := range state {
		state[i] = 0
	}
	h := GnarkMDHasher{api: api, state: state}
	h.Write(c.Inputs...)
	res := h.Sum()
	for i := 0; i < 8; i++ {
		api.AssertIsEqual(c.Output[i], res[i])
	}
	return nil
}

func getNativeMDHasherCircuitWitness(nbElmts int) (*nativeMDHasherCircuit, *nativeMDHasherCircuit) {
	vals := make([]field.Element, nbElmts)
	for i := 0; i < nbElmts; i++ {
		vals[i].SetRandom()
	}
	phasher := NewMDHasher()
	phasher.WriteElements(vals...)
	res := phasher.SumElement()

	var circuit, witness nativeMDHasherCircuit
	circuit.Inputs = make([]frontend.Variable, nbElmts)
	witness.Inputs = make([]frontend.Variable, nbElmts)
	for i := 0; i < nbElmts; i++ {
		witness.Inputs[i] = vals[i].String()
	}
	for i := 0; i < 8; i++ {
		witness.Output[i] = res[i].String()
	}
	return &circuit, &witness
}

// TestCircuit tests the native (non-GKR) Poseidon2 gnark circuit.
// Uses scs.NewBuilder directly — no wideCommitWrapper, no RegisterGates.
//
// Compare with BLS12-377:
//
//	go test ./crypto/poseidon2_bls12377/... -run TestCircuit -v
//	go test ./crypto/poseidon2_koalabear/... -run "^TestCircuit$" -v
func TestNativeCircuit(t *testing.T) {
	for _, size := range []int{8, 16, 444, 800, 160000} {
		size := size
		t.Run("Size_"+strconv.Itoa(size), func(t *testing.T) {
			circuit, witness := getNativeMDHasherCircuitWitness(size)

			t0 := time.Now()
			ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, circuit)
			assert.NoError(t, err)
			compileElapsed := time.Since(t0)

			fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
			assert.NoError(t, err)

			t1 := time.Now()
			err = ccs.IsSolved(fullWitness)
			assert.NoError(t, err)
			solveElapsed := time.Since(t1)

			fmt.Printf("[native] size=%d constraints=%d compile=%s solve=%s\n",
				size, ccs.GetNbConstraints(), compileElapsed, solveElapsed)
		})
	}
}
