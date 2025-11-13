package fiatshamir

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/stretchr/testify/assert"
)

type FSCircuit struct {
	A, B   frontend.Variable
	R1, R2 poseidon2_koalabear.Octuplet
}

func (c *FSCircuit) Define(api frontend.API) error {
	fs := NewGnarkFS(api)
	fs.Update(c.A)
	a := fs.RandomField()
	fs.Update(c.B)
	b := fs.RandomField()
	for i := 0; i < 8; i++ {
		api.AssertIsEqual(a[i], c.R1[i])
		api.AssertIsEqual(b[i], c.R2[i])
	}
	return nil
}

func GetCircuitWitnessFSCircuit() (*FSCircuit, *FSCircuit) {

	h := poseidon2_koalabear.NewMDHasher()
	var a, b koalabear.Element
	a.SetRandom()
	b.SetRandom()
	Update(h, a)
	r1 := RandomField(h)
	Update(h, b)
	r2 := RandomField(h)

	var circuit, witness FSCircuit
	witness.A = a.String()
	witness.B = b.String()
	for i := 0; i < 8; i++ {
		witness.R1[i] = r1[i].String()
		witness.R2[i] = r2[i].String()
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
