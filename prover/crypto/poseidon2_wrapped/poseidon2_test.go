package poseidon2_wrapped

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/stretchr/testify/assert"
)

type GnarkHasherCircuit struct {
	Inputs []zk.WrappedVariable
	Ouput  GHash
}

func (ghc *GnarkHasherCircuit) Define(api frontend.API) error {

	h, err := NewGnarkHasher(api)
	if err != nil {
		return err
	}

	// write elmts
	h.Write(ghc.Inputs...)

	// sum
	res := h.Sum()

	// check the result
	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		return err
	}
	for i := 0; i < 8; i++ {
		apiGen.AssertIsEqual(ghc.Ouput[i], res[i])
	}

	return nil
}

func getGnarkHasherCircuitWitness() (*GnarkHasherCircuit, *GnarkHasherCircuit) {

	// values to hash
	nbElmts := 2
	vals := make([]field.Element, nbElmts)
	for i := 0; i < nbElmts; i++ {
		// vals[i].SetRandom()
		vals[i].SetUint64(uint64(10 + i))
	}

	// sum
	phasher := poseidon2.Poseidon2()
	phasher.WriteElements(vals)
	res := phasher.SumElement()

	// create witness and circuit
	var circuit, witness GnarkHasherCircuit
	circuit.Inputs = make([]zk.WrappedVariable, nbElmts)
	witness.Inputs = make([]zk.WrappedVariable, nbElmts)
	for i := 0; i < nbElmts; i++ {
		witness.Inputs[i] = zk.ValueOf(vals[i].String())
	}
	for i := 0; i < 8; i++ {
		witness.Ouput[i] = zk.ValueOf(res[i].String())
	}

	return &circuit, &witness

}

func TestCircuit(t *testing.T) {

	{
		circuit, witness := getGnarkHasherCircuitWitness()

		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

	{
		circuit, witness := getGnarkHasherCircuitWitness()

		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

}
