package poseidon2_koalabear

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
	"github.com/stretchr/testify/assert"
)

//---------------------------------------
// wrapped variables

type GnarkMDHasherCircuitWV struct {
	Inputs []koalagnark.Element
	Ouput  koalagnark.Octuplet
}

func (ghc *GnarkMDHasherCircuitWV) Define(api frontend.API) error {

	h, err := NewGnarkMDHasherWV(api)
	if err != nil {
		return err
	}

	// write elmts
	h.Write(ghc.Inputs...)

	// sum
	res := h.Sum()

	// check the result
	koalaAPI := koalagnark.NewAPI(api)

	for i := 0; i < 8; i++ {
		koalaAPI.AssertIsEqual(ghc.Ouput[i], res[i])
	}

	return nil
}

func getGnarkMDHasherCircuitWitnessWV() (*GnarkMDHasherCircuitWV, *GnarkMDHasherCircuitWV) {

	// values to hash
	nbElmts := 16
	vals := make([]field.Element, nbElmts)
	for i := 0; i < nbElmts; i++ {
		vals[i].SetRandom()
		// vals[i].SetUint64(uint64(10 + i))
	}

	// sum
	phasher := NewMDHasher()
	phasher.WriteElements(vals...)
	res := phasher.SumElement()

	// create witness and circuit
	var circuit, witness GnarkMDHasherCircuitWV
	circuit.Inputs = make([]koalagnark.Element, nbElmts)
	witness.Inputs = make([]koalagnark.Element, nbElmts)
	for i := 0; i < nbElmts; i++ {
		witness.Inputs[i] = koalagnark.NewElementFromKoala(vals[i])
	}
	for i := 0; i < 8; i++ {
		witness.Ouput[i] = koalagnark.NewElementFromKoala(res[i])
	}

	return &circuit, &witness

}

func TestCircuitWV(t *testing.T) {

	circuit, witness := getGnarkMDHasherCircuitWitnessWV()

	ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, circuit)
	assert.NoError(t, err)

	fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
	assert.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err)

}

//---------------------------------------
// native variables

type GnarkMDHasherCircuit struct {
	Inputs []frontend.Variable
	Ouput  Octuplet
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

func getGnarkMDHasherCircuitWitness() (*GnarkMDHasherCircuit, *GnarkMDHasherCircuit) {

	// values to hash
	nbElmts := 16
	vals := make([]field.Element, nbElmts)
	for i := 0; i < nbElmts; i++ {
		vals[i].SetRandom()
		// vals[i].SetUint64(uint64(10 + i))
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

func TestCircuit(t *testing.T) {

	circuit, witness := getGnarkMDHasherCircuitWitness()

	ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, circuit)
	assert.NoError(t, err)

	fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
	assert.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err)

}
