package fiatshamir_bls12377

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

type FSCircuit struct {
	// frElements
	A, B   frontend.Variable
	R1, R2 frontend.Variable

	// koalabear octuplet
	C  [2]zk.WrappedVariable
	D  [10]zk.WrappedVariable
	R3 zk.Octuplet

	// random many integers
	R4    []frontend.Variable
	n     int
	bound int

	// set state, get state
	SetState, GetState zk.Octuplet
}

func (c *FSCircuit) Define(api frontend.API) error {

	fs := NewGnarkFS(api)

	// frElements
	fs.UpdateFrElmt(c.A)
	a := fs.RandomFrElmt()
	fs.UpdateFrElmt(c.B)
	b := fs.RandomFrElmt()
	api.AssertIsEqual(a, c.R1)
	api.AssertIsEqual(b, c.R2)

	// koalabear octuplet
	fs.Update(c.C[:]...)
	e := fs.RandomField()
	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		return err
	}
	for i := 0; i < 8; i++ {
		apiGen.AssertIsEqual(e[i], c.R3[i])
	}

	// random many integers
	fs.Update(c.D[:]...)
	res := fs.RandomManyIntegers(c.n, c.bound)
	for i := 0; i < len(res); i++ {
		api.AssertIsEqual(res[i], c.R4[i])
	}

	// set state, get state
	fs.SetState(c.SetState)
	getState := fs.State()
	for i := 0; i < len(getState); i++ {
		apiGen.AssertIsEqual(getState[i], c.GetState[i])
	}

	return nil
}

func GetCircuitWitnessFSCircuit() (*FSCircuit, *FSCircuit) {

	fs := NewFS()

	// fr element
	var a, b fr.Element
	a.SetRandom()
	b.SetRandom()
	fs.UpdateFrElmt(a)
	r1 := fs.RandomFieldFrElmt()
	fs.UpdateFrElmt(b)
	r2 := fs.RandomFieldFrElmt()

	// koalabear element
	var c [2]field.Element
	c[0].SetRandom()
	c[1].SetRandom()
	var d [10]field.Element
	for i := 0; i < 10; i++ {
		d[i].SetRandom()
	}
	fs.Update(c[:]...)
	r3 := fs.RandomField()

	// random many integers
	fs.Update(d[:]...)
	n := 5
	bound := 8
	r4 := fs.RandomManyIntegers(n, bound)

	// set state, get state
	var setState field.Octuplet
	for i := 0; i < 8; i++ {
		setState[i].SetRandom()
	}
	fs.SetState(setState)
	getState := fs.State()

	var circuit, witness FSCircuit
	witness.A = a.String()
	witness.B = b.String()
	witness.R1 = r1.String()
	witness.R2 = r2.String()
	witness.C[0] = zk.ValueFromKoala(c[0])
	witness.C[1] = zk.ValueFromKoala(c[1])
	for i := 0; i < 10; i++ {
		witness.D[i] = zk.ValueFromKoala(d[i])
	}
	for i := 0; i < 8; i++ {
		witness.R3[i] = zk.ValueFromKoala(r3[i])
		witness.SetState[i] = zk.ValueFromKoala(setState[i])
		witness.GetState[i] = zk.ValueFromKoala(getState[i])
	}

	circuit.n = n
	circuit.bound = bound
	witness.R4 = make([]frontend.Variable, n)
	circuit.R4 = make([]frontend.Variable, n)
	for i := 0; i < n; i++ {
		witness.R4[i] = r4[i]
	}

	return &circuit, &witness
}

func TestFSCircuit(t *testing.T) {

	circuit, witness := GetCircuitWitnessFSCircuit()
	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circuit)
	assert.NoError(t, err)

	fullWitness, err := frontend.NewWitness(witness, ecc.BLS12_377.ScalarField())
	assert.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err)
}

type KoalaFlushSpecificCircuit struct {
	Input   [32]zk.WrappedVariable
	Output  zk.Octuplet
	isKoala bool
}

func (c *KoalaFlushSpecificCircuit) Define(api frontend.API) error {
	fs := NewGnarkFS(api)

	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		return err
	}

	fs.Update(c.Input[:]...)
	res := fs.RandomField()
	for i := 0; i < len(res); i++ {
		apiGen.AssertIsEqual(res[i], c.Output[i])
	}
	return nil
}

func getKoalaFlushSpecificWitness(isKoala bool) (*KoalaFlushSpecificCircuit, *KoalaFlushSpecificCircuit) {
	var circuit, witness KoalaFlushSpecificCircuit
	circuit.isKoala = isKoala
	witness.isKoala = isKoala
	fs := NewFS()

	var input [32]field.Element
	var one field.Element
	one.SetOne()
	// zero is 0 by default

	for i := 0; i < 16; i++ {
		// input[2*i] = -1
		input[2*i] = *field.MaxVal
		// input[2*i+1] = 0
		input[2*i+1].SetZero()
	}

	fs.Update(input[:]...)
	output := fs.RandomField()

	for i := 0; i < 32; i++ {
		witness.Input[i] = zk.ValueFromKoala(input[i])
	}
	for i := 0; i < 8; i++ {
		witness.Output[i] = zk.ValueFromKoala(output[i])
	}

	return &circuit, &witness
}

func TestKoalaFlushSpecific(t *testing.T) {

	// compile on bls
	{
		circuit, witness := getKoalaFlushSpecificWitness(false)
		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}
}
