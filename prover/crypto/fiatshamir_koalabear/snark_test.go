package fiatshamir

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/stretchr/testify/assert"
)

type FSCircuit struct {
	A, B   frontend.Variable
	R1, R2 gnarkfext.E4Gen
}

func (c *FSCircuit) Define(api frontend.API) error {
	fs := NewGnarkFS(api)
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
	return nil
}

func GetCircuitWitnessFSCircuit() (*FSCircuit, *FSCircuit) {

	h := poseidon2_koalabear.NewMDHasher()
	var a, b koalabear.Element
	a.SetRandom()
	b.SetRandom()
	Update(h, a)
	r1 := RandomFext(h)
	Update(h, b)
	r2 := RandomFext(h)

	var circuit, witness FSCircuit
	witness.A = a.String()
	witness.B = b.String()

	witness.R1 = gnarkfext.NewE4Gen(r1)
	witness.R2 = gnarkfext.NewE4Gen(r2)

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
