package fiatshamir

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir_bls12377"
	"github.com/consensys/linea-monorepo/prover/crypto/fiatshamir_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/stretchr/testify/assert"
)

type FSCircuit struct {
	A                  [10]zk.WrappedVariable
	RandomA            zk.Octuplet
	B                  [10]gnarkfext.E4Gen
	RandomB            zk.Octuplet
	RandomField        zk.Octuplet
	RandomFieldExt     gnarkfext.E4Gen
	RandomManyIntegers [10]frontend.Variable
	isKoala            bool
}

func assertEqualOctuplet(api zk.GenericApi, a, b zk.Octuplet) {
	for i := 0; i < 8; i++ {
		api.AssertIsEqual(a[i], b[i])
	}
}

func (c *FSCircuit) Define(api frontend.API) error {

	var fs GnarkFS
	if c.isKoala {
		fs = NewGnarkFSKoalabear(api)
	} else {
		fs = NewGnarkFSKoalaBLS12377(api)
	}

	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		return err
	}

	fs.Update(c.A[:]...)
	randomA := fs.RandomField()
	assertEqualOctuplet(apiGen, randomA, c.RandomA)

	fs.UpdateExt(c.B[:]...)
	randomB := fs.RandomField()
	assertEqualOctuplet(apiGen, randomB, c.RandomB)

	randomField := fs.RandomField()

	assertEqualOctuplet(apiGen, randomField, c.RandomField)

	randomFieldExt := fs.RandomFieldExt()
	apiGen.AssertIsEqual(randomFieldExt.B0.A0, c.RandomFieldExt.B0.A0)
	apiGen.AssertIsEqual(randomFieldExt.B0.A1, c.RandomFieldExt.B0.A1)
	apiGen.AssertIsEqual(randomFieldExt.B1.A0, c.RandomFieldExt.B1.A0)
	apiGen.AssertIsEqual(randomFieldExt.B1.A1, c.RandomFieldExt.B1.A1)

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
		witness.A[i] = zk.ValueOf(A[i].String())
	}
	for i := 0; i < 8; i++ {
		witness.RandomA[i] = zk.ValueOf(RandomA[i].String())
	}

	for i := 0; i < 10; i++ {
		witness.B[i].B0.A0 = zk.ValueOf(B[i].B0.A0.String())
		witness.B[i].B0.A1 = zk.ValueOf(B[i].B0.A1.String())
		witness.B[i].B1.A0 = zk.ValueOf(B[i].B1.A0.String())
		witness.B[i].B1.A1 = zk.ValueOf(B[i].B1.A1.String())
	}
	for i := 0; i < 8; i++ {
		witness.RandomB[i] = zk.ValueOf(RandomB[i].String())
	}

	for i := 0; i < 8; i++ {
		witness.RandomField[i] = zk.ValueOf(RandomField[i].String())
	}

	witness.RandomFieldExt.B0.A0 = zk.ValueOf(RandomFieldExt.B0.A0.String())
	witness.RandomFieldExt.B0.A1 = zk.ValueOf(RandomFieldExt.B0.A1.String())
	witness.RandomFieldExt.B1.A0 = zk.ValueOf(RandomFieldExt.B1.A0.String())
	witness.RandomFieldExt.B1.A1 = zk.ValueOf(RandomFieldExt.B1.A1.String())

	for i := 0; i < 10; i++ {
		witness.RandomManyIntegers[i] = RandomManyIntegers[i]
	}

	return &circuit, &witness
}

func TestFSCircuit(t *testing.T) {

	// compile on koala
	{
		getWitnessCircuit(true)
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
		getWitnessCircuit(true)
		circuit, witness := getWitnessCircuit(false)
		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

}
