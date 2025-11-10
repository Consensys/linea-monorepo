package poseidon2_koalabear

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/stretchr/testify/assert"
)

type MerkleDamgardHasherCircuit struct {
	Inputs []frontend.Variable
	Ouput  Octuplet
}

func (ghc *MerkleDamgardHasherCircuit) Define(api frontend.API) error {

	h, err := NewMerkleDamgardHasher(api)
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

func getMerkleDamgardHasherCircuitWitness() (*MerkleDamgardHasherCircuit, *MerkleDamgardHasherCircuit) {

	// values to hash
	nbElmts := 2
	vals := make([]field.Element, nbElmts)
	for i := 0; i < nbElmts; i++ {
		// vals[i].SetRandom()
		vals[i].SetUint64(uint64(10 + i))
	}

	// sum
	phasher := Poseidon2()
	phasher.WriteElements(vals)
	res := phasher.SumElement()

	// create witness and circuit
	var circuit, witness MerkleDamgardHasherCircuit
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

func TestCircuit(t *testing.T) {

	circuit, witness := getMerkleDamgardHasherCircuitWitness()

	ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, circuit)
	assert.NoError(t, err)

	fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
	assert.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err)

}
