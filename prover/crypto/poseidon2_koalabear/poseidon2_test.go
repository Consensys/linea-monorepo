package poseidon2_koalabear

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	_ "github.com/consensys/gnark/std/hash/all"
	"github.com/consensys/linea-monorepo/prover/maths/field"
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
func TestCircuitSolve(t *testing.T) {
	require.NoError(t, RegisterGates())

	for _, nbElmts := range []int{8, 16, 444} {
		t.Run(fmt.Sprintf("Size_%d", nbElmts), func(t *testing.T) {
			circuit, witness := getGnarkMDHasherCircuitWitness(nbElmts)

			ccs, err := frontend.CompileU32(koalabear.Modulus(), wideCommitWrapper(scs.NewBuilder), circuit)
			require.NoError(t, err)

			fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
			require.NoError(t, err)
			err = ccs.IsSolved(fullWitness)
			require.NoError(t, err)
			fmt.Printf("GKR Poseidon2 Solve (%d elems, %d instances): %d constraints\n",
				nbElmts, (nbElmts+7)/8, ccs.GetNbConstraints())
		})
	}
}
