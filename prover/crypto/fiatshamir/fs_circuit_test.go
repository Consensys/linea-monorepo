package fiatshamir

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir_bls12377"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir_koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/stretchr/testify/assert"
)

type FSCircuit struct {
	A                  [10]frontend.Variable
	RandomA            poseidon2_koalabear.Octuplet
	B                  [10]gnarkfext.Element
	RandomB            poseidon2_koalabear.Octuplet
	RandomField        poseidon2_koalabear.Octuplet
	RandomFieldExt     gnarkfext.Element
	RandomManyIntegers [10]frontend.Variable
	isKoala            bool
}

func assertEqualOctuplet(api frontend.API, a, b poseidon2_koalabear.Octuplet) {
	for i := 0; i < 8; i++ {
		api.AssertIsEqual(a[i], b[i])
	}
}

func (c *FSCircuit) Define(api frontend.API) error {

	var fs GnarkFS
	if c.isKoala {
		fs = NewGnarkFSKoalabear(api)
	} else {
		fs = NewGnarkFSBLS12377(api)
	}

	fs.Update(c.A[:]...)
	randomA := fs.RandomField()
	assertEqualOctuplet(api, randomA, c.RandomA)

	fs.UpdateExt(c.B[:]...)
	randomB := fs.RandomField()
	assertEqualOctuplet(api, randomB, c.RandomB)

	randomField := fs.RandomField()

	assertEqualOctuplet(api, randomField, c.RandomField)

	randomFieldExt := fs.RandomFieldExt()

	gnarkfext.AssertIsEqual(api, randomFieldExt, c.RandomFieldExt)

	randomManyIntegers := fs.RandomManyIntegers(10, 16)
	for i := 0; i < 10; i++ {
		api.AssertIsEqual(randomManyIntegers[i], c.RandomManyIntegers[i])
	}

	return nil
}

func getWitnessCircuit(isKoala bool) (*FSCircuit, *FSCircuit) {

	var fs FS
	var circuit, witness FSCircuit
	circuit.isKoala = isKoala
	witness.isKoala = isKoala
	if isKoala {
		fs = fiatshamir_koalabear.NewFS()
	} else {
		fs = fiatshamir_bls12377.NewFS()
	}

	var A [10]field.Element
	for i := 0; i < 10; i++ {
		A[i].SetRandom()
	}
	fs.Update(A[:]...)
	RandomA := fs.RandomField()

	var B [10]fext.Element
	for i := 0; i < 10; i++ {
		B[i].SetRandom()
	}
	fs.UpdateExt(B[:]...)

	RandomB := fs.RandomField()

	RandomField := fs.RandomField()
	RandomFieldExt := fs.RandomFext()
	RandomManyIntegers := fs.RandomManyIntegers(10, 16)

	for i := 0; i < 10; i++ {
		witness.A[i] = field.NewFrontendFromKoala(A[i])
	}
	for i := 0; i < 8; i++ {
		witness.RandomA[i] = field.NewFrontendFromKoala(RandomA[i])
	}

	for i := 0; i < 10; i++ {
		witness.B[i] = gnarkfext.AssignFromExt(B[i])
	}
	for i := 0; i < 8; i++ {
		witness.RandomB[i] = field.NewFrontendFromKoala(RandomB[i])
	}

	for i := 0; i < 8; i++ {
		witness.RandomField[i] = field.NewFrontendFromKoala(RandomField[i])
	}

	witness.RandomFieldExt = gnarkfext.AssignFromExt(RandomFieldExt)

	for i := 0; i < 10; i++ {
		witness.RandomManyIntegers[i] = RandomManyIntegers[i]
	}

	return &circuit, &witness
}

func TestFSCircuit(t *testing.T) {

	// compile on koala
	{
		circuit, witness := getWitnessCircuit(true)
		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

	// compile on bls
	{
		circuit, witness := getWitnessCircuit(false)
		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

}
