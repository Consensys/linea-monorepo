package fiatshamir_bls12377

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_bls12377"
	"github.com/stretchr/testify/assert"
)

type FSCircuit struct {
	A, B   frontend.Variable
	R1, R2 frontend.Variable
}

func (c *FSCircuit) Define(api frontend.API) error {
	fs := NewGnarkFS(api)
	fs.UpdateFrElmt(c.A)
	a := fs.RandomField()
	fs.UpdateFrElmt(c.B)
	b := fs.RandomField()
	api.Println(b)
	api.Println(c.R2)
	api.AssertIsEqual(a, c.R1)
	api.AssertIsEqual(b, c.R2)
	return nil
}

func GetCircuitWitnessFSCircuit() (*FSCircuit, *FSCircuit) {

	h := poseidon2_bls12377.NewMDHasher()
	var a, b fr.Element
	a.SetRandom()
	b.SetRandom()
	UpdateFrElmt(h, a)
	r1 := RandomFieldFrElmt(h)
	UpdateFrElmt(h, b)
	r2 := RandomFieldFrElmt(h)

	var circuit, witness FSCircuit
	witness.A = a.String()
	witness.B = b.String()
	witness.R1 = r1.String()
	witness.R2 = r2.String()
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
