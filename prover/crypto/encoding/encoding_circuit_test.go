package encoding

import (
	"errors"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/profile"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/stretchr/testify/assert"
)

type EncodingCircuit struct {
	ToEncode1 [8]zk.WrappedVariable
	ToEncode2 [12]zk.WrappedVariable
	ToEncode3 frontend.Variable
	R1        frontend.Variable
	R2        [2]frontend.Variable
	R3        zk.Octuplet
}

func (c *EncodingCircuit) Define(api frontend.API) error {

	a := EncodeWVsToFVs(api, c.ToEncode1[:])
	b := EncodeWVsToFVs(api, c.ToEncode2[:])
	d := EncodeFVTo8WVs(api, c.ToEncode3)
	if len(a) != 1 {
		return errors.New("ToEncode1 should correspond to a single frelement")
	}
	if len(b) != 2 {
		return errors.New("ToEncode2should correspond to 2 frelement")
	}

	api.AssertIsEqual(a[0], c.R1)
	api.AssertIsEqual(b[0], c.R2[0])
	api.AssertIsEqual(b[1], c.R2[1])

	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		return err
	}
	for i := 0; i < 8; i++ {
		apiGen.AssertIsEqual(c.R3[i], d[i])
	}

	return nil
}

func TestEncoding(t *testing.T) {

	// get witness
	var witness EncodingCircuit
	var toEncode1 [8]field.Element
	for i := 0; i < 8; i++ {
		toEncode1[i].SetRandom()
		witness.ToEncode1[i] = zk.ValueOf(toEncode1[i].String())
	}
	var toEncode2 [12]field.Element
	for i := 0; i < 12; i++ {
		toEncode2[i].SetRandom()
		witness.ToEncode2[i] = zk.ValueOf(toEncode2[i].String())
	}
	var toEncode3 fr.Element
	toEncode3.SetRandom()
	witness.ToEncode3 = toEncode3.String()

	r1 := EncodeKoalabearsToFrElement(toEncode1[:])
	witness.R1 = r1[0].String()
	r2 := EncodeKoalabearsToFrElement(toEncode2[:])
	witness.R2[0] = r2[0].String()
	witness.R2[1] = r2[1].String()
	r3 := EncodeFrElementToOctuplet(toEncode3)
	for i := 0; i < 8; i++ {
		witness.R3[i] = zk.ValueOf(r3[i].String())
	}

	var circuit EncodingCircuit
	filePath := "TestEncoding.pprof"
	pro := profile.Start(profile.WithPath(filePath))

	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit)
	pro.Stop()

	assert.NoError(t, err)
	fullWitness, err := frontend.NewWitness(&witness, ecc.BLS12_377.ScalarField())
	assert.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err)

}
