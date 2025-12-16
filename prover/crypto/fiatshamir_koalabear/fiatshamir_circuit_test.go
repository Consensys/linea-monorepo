package fiatshamir_koalabear

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/stretchr/testify/assert"
)

type FSCircuit struct {

	// random fext
	A, B   frontend.Variable
	R1, R2 gnarkfext.E4Gen

	// random many integers
	R3       []frontend.Variable
	n, bound int

	// set state, get state
	SetState, GetState poseidon2_koalabear.Octuplet
}

func (c *FSCircuit) Define(api frontend.API) error {
	fs := NewGnarkFS(api)

	// random fext
	fs.Update(c.A)
	a := fs.RandomFieldExt()
	fs.Update(c.B)
	b := fs.RandomFieldExt()
	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		return err
	}
	ext4.AssertIsEqual(&a, &c.R1)
	ext4.AssertIsEqual(&b, &c.R2)

	// random many integers
	r := fs.RandomManyIntegers(c.n, c.bound)
	for i := 0; i < c.n; i++ {
		api.AssertIsEqual(r[i], c.R3[i])
	}

	// set state, get state
	fs.SetState(c.SetState)
	getState := fs.State()
	for i := 0; i < len(getState); i++ {
		api.AssertIsEqual(getState[i], c.GetState[i])
	}

	return nil
}

func GetCircuitWitnessFSCircuit() (*FSCircuit, *FSCircuit) {

	fs := NewFS()
	var a, b koalabear.Element
	a.SetRandom()
	b.SetRandom()
	fs.Update(a)
	r1 := fs.RandomFext()
	fs.Update(b)
	r2 := fs.RandomFext()

	n := 4
	bound := 8
	r3 := fs.RandomManyIntegers(n, bound)

	var setState field.Octuplet
	for i := 0; i < 8; i++ {
		setState[i].SetRandom()
	}
	fs.SetState(setState)
	getSate := fs.State()

	var circuit, witness FSCircuit
	circuit.n = n
	circuit.bound = bound
	circuit.R3 = make([]frontend.Variable, n)
	witness.A = a.String()
	witness.B = b.String()
	witness.R1 = gnarkfext.NewE4Gen(r1)
	witness.R2 = gnarkfext.NewE4Gen(r2)
	witness.R3 = make([]frontend.Variable, n)
	for i := 0; i < n; i++ {
		witness.R3[i] = r3[i]
	}
	for i := 0; i < 8; i++ {
		witness.SetState[i] = setState[i]
		witness.GetState[i] = getSate[i]
	}

	return &circuit, &witness
}

func TestFSCircuit(t *testing.T) {

	circuit, witness := GetCircuitWitnessFSCircuit()
	ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, circuit)
	assert.NoError(t, err)

	fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
	assert.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err)
}
