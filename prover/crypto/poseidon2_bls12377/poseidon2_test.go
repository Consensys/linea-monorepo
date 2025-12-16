package poseidon2_bls12377

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/stretchr/testify/assert"
)

// ------------------------------------------------------------
// hashing bls12377 elmts

type GnarkMDHasherCircuit struct {
	Inputs []frontend.Variable
	Ouput  frontend.Variable
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
	api.AssertIsEqual(ghc.Ouput, res)

	return nil
}

func getGnarkMDHasherCircuitWitness() (*GnarkMDHasherCircuit, *GnarkMDHasherCircuit) {

	// values to hash
	nbElmts := 2025 // TODO@yao: test with different sizes
	vals := make([]fr.Element, nbElmts)
	for i := 0; i < nbElmts; i++ {
		// vals[i].SetRandom()
		vals[i].SetUint64(uint64(10 + i))
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
	witness.Ouput = res.String()

	return &circuit, &witness

}

func TestCircuit(t *testing.T) {

	circuit, witness := getGnarkMDHasherCircuitWitness()

	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circuit)
	assert.NoError(t, err)

	fullWitness, err := frontend.NewWitness(witness, ecc.BLS12_377.ScalarField())
	assert.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err)

}

// ------------------------------------------------------------
// hashing koalabear elmts

type GnarkMDHasherCircuitKoalabear struct {
	Inputs []zk.WrappedVariable
	Ouput  frontend.Variable
}

func (ghc *GnarkMDHasherCircuitKoalabear) Define(api frontend.API) error {

	h, err := NewGnarkMDHasher(api)
	if err != nil {
		return err
	}

	// write elmts
	h.WriteWVs(ghc.Inputs...)

	// sum
	res := h.Sum()

	// check the result
	api.AssertIsEqual(ghc.Ouput, res)

	return nil
}

func getGnarkMDHasherCircuitKoalabearWitness() (*GnarkMDHasherCircuitKoalabear, *GnarkMDHasherCircuitKoalabear) {

	// values to hash
	nbElmts := 4096
	vals := make([]field.Element, nbElmts)
	for i := 0; i < nbElmts; i++ {
		// vals[i].SetRandom()
		vals[i].SetUint64(uint64(10 + i))
	}

	// sum
	phasher := NewMDHasher()
	phasher.WriteKoalabearElements(vals...)
	res := phasher.SumElement()

	// create witness and circuit
	var circuit, witness GnarkMDHasherCircuitKoalabear
	circuit.Inputs = make([]zk.WrappedVariable, nbElmts)
	witness.Inputs = make([]zk.WrappedVariable, nbElmts)
	for i := 0; i < nbElmts; i++ {
		witness.Inputs[i] = zk.ValueOf(vals[i].String())
	}
	witness.Ouput = res.String()

	return &circuit, &witness

}

func TestCircuitKoalabear(t *testing.T) {

	circuit, witness := getGnarkMDHasherCircuitKoalabearWitness()

	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circuit)
	assert.NoError(t, err)

	fullWitness, err := frontend.NewWitness(witness, ecc.BLS12_377.ScalarField())
	assert.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err)

}
