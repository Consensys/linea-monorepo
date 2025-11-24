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
	R4 [2]zk.Octuplet
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
	fs.UpdateElmts(c.C[:]...)
	e := fs.RandomField()
	fs.UpdateElmts(c.D[:]...)
	f := fs.RandomManyIntegers(2)
	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		return err
	}
	for i := 0; i < 8; i++ {
		apiGen.AssertIsEqual(e[i], c.R3[i])
	}
	for i := 0; i < len(f); i++ {
		for j := 0; j < 8; j++ {
			apiGen.AssertIsEqual(f[i][j], c.R4[i][j])
		}
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
	fs.UpdateElmts(c[:]...)
	r3 := fs.RandomField()
	fs.UpdateElmts(d[:]...)
	r4 := fs.RandomManyIntegers(2)

	var circuit, witness FSCircuit
	witness.A = a.String()
	witness.B = b.String()
	witness.R1 = r1.String()
	witness.R2 = r2.String()
	witness.C[0] = zk.ValueOf(c[0].String())
	witness.C[1] = zk.ValueOf(c[1].String())
	for i := 0; i < 10; i++ {
		witness.D[i] = zk.ValueOf(d[i].String())
	}
	for i := 0; i < 8; i++ {
		witness.R3[i] = zk.ValueOf(r3[i].String())
	}
	for i := 0; i < 2; i++ {
		for j := 0; j < 8; j++ {
			witness.R4[i][j] = zk.ValueOf(r4[i][j].String())
		}
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
