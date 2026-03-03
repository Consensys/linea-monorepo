package fiatshamir_koalabear

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// KoalagnarkConsistencyCircuit tests that GnarkFSKoalagnark produces the same
// output as GnarkFSWV. Both implementations use different code paths but
// should produce identical Poseidon2 hash results.
type KoalagnarkConsistencyCircuit struct {
	Input [2]koalagnark.Element
	// Expected output from native GnarkFSWV
	ExpectedNative koalagnark.Octuplet
}

func (c *KoalagnarkConsistencyCircuit) Define(api frontend.API) error {
	// Use GnarkFSKoalagnark
	fs := NewGnarkFSKoalagnark(api)
	fs.Update(c.Input[:]...)
	res := fs.RandomField()

	koalaAPI := koalagnark.NewAPI(api)
	for i := 0; i < 8; i++ {
		koalaAPI.AssertIsEqual(res[i], c.ExpectedNative[i])
	}
	return nil
}

// TestKoalagnarkConsistencyWithWVOnNative verifies that GnarkFSKoalagnark and
// GnarkFSWV produce the same output when both run on native KoalaBear field.
// This isolates the hasher consistency from emulation concerns.
func TestKoalagnarkConsistencyWithWVOnNative(t *testing.T) {
	// Compute expected output using non-circuit FS
	nativeFS := NewFS()
	var input [2]field.Element
	input[0].SetRandom()
	input[1].SetRandom()
	nativeFS.Update(input[:]...)
	output := nativeFS.RandomField()

	// Create witness
	var circuit KoalagnarkConsistencyCircuit
	witness := KoalagnarkConsistencyCircuit{
		Input: [2]koalagnark.Element{
			koalagnark.NewElement(input[0].Uint64()),
			koalagnark.NewElement(input[1].Uint64()),
		},
	}
	for i := 0; i < 8; i++ {
		witness.ExpectedNative[i] = koalagnark.NewElement(output[i].Uint64())
	}

	// Compile on KoalaBear native first
	ccs, err := frontend.CompileU32(koalabear.Modulus(), wideCommitWrapper(scs.NewBuilder), &circuit)
	require.NoError(t, err, "compilation on KoalaBear should succeed")
	t.Logf("Native KoalaBear FS circuit: %d constraints", ccs.GetNbConstraints())

	fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
	require.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err, "GnarkFSKoalagnark should match non-circuit FS on native KoalaBear")
}

// TestKoalagnarkUpdateExtConsistencyNative tests UpdateExt consistency on native field.
func TestKoalagnarkUpdateExtConsistencyNative(t *testing.T) {
	nativeFS := NewFS()
	var extInputs [2]fext.Element
	for i := 0; i < 2; i++ {
		extInputs[i].SetRandom()
	}
	nativeFS.UpdateExt(extInputs[:]...)
	output := nativeFS.RandomField()

	var circuit, witness KoalagnarkExtConsistencyCircuit
	for i := 0; i < 2; i++ {
		witness.ExtInputs[i] = koalagnark.NewExt(extInputs[i])
	}
	for i := 0; i < 8; i++ {
		witness.ExpectedNative[i] = koalagnark.NewElement(output[i].Uint64())
	}

	ccs, err := frontend.CompileU32(koalabear.Modulus(), wideCommitWrapper(scs.NewBuilder), &circuit)
	require.NoError(t, err)
	t.Logf("Native KoalaBear ext FS circuit: %d constraints", ccs.GetNbConstraints())

	fullWitness, err := frontend.NewWitness(&witness, koalabear.Modulus())
	require.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err, "GnarkFSKoalagnark UpdateExt should match non-circuit FS on native KoalaBear")
}

type KoalagnarkExtConsistencyCircuit struct {
	ExtInputs      [2]koalagnark.Ext
	ExpectedNative koalagnark.Octuplet
}

func (c *KoalagnarkExtConsistencyCircuit) Define(api frontend.API) error {
	fs := NewGnarkFSKoalagnark(api)
	fs.UpdateExt(c.ExtInputs[:]...)
	res := fs.RandomField()
	koalaAPI := koalagnark.NewAPI(api)
	for i := 0; i < 8; i++ {
		koalaAPI.AssertIsEqual(res[i], c.ExpectedNative[i])
	}
	return nil
}

// TestKoalagnarkFSOnBN254 validates that GnarkFSKoalagnark produces the
// same Fiat-Shamir output as the non-circuit FS when compiled on BN254.
// This is the critical test for the emulated mode.
type KoalagnarkBN254Circuit struct {
	Input          [2]koalagnark.Element
	ExpectedNative koalagnark.Octuplet
}

func (c *KoalagnarkBN254Circuit) Define(api frontend.API) error {
	fs := NewGnarkFSKoalagnark(api)
	fs.Update(c.Input[:]...)
	res := fs.RandomField()

	koalaAPI := koalagnark.NewAPI(api)
	for i := 0; i < 8; i++ {
		koalaAPI.AssertIsEqual(res[i], c.ExpectedNative[i])
	}
	return nil
}

func TestKoalagnarkFSOnBN254(t *testing.T) {
	// Step 1: Compute expected output using non-circuit FS
	fs := NewFS()
	var input [2]field.Element
	input[0].SetRandom()
	input[1].SetRandom()
	fs.Update(input[:]...)
	output := fs.RandomField()

	// Step 2: Create circuit + witness
	var circuit KoalagnarkBN254Circuit
	witness := KoalagnarkBN254Circuit{
		Input: [2]koalagnark.Element{
			koalagnark.NewElement(input[0].Uint64()),
			koalagnark.NewElement(input[1].Uint64()),
		},
	}
	for i := 0; i < 8; i++ {
		witness.ExpectedNative[i] = koalagnark.NewElement(output[i].Uint64())
	}

	// Step 3: Compile on BN254 (emulated mode!)
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, &circuit)
	require.NoError(t, err, "compilation on BN254 should succeed")
	t.Logf("BN254 emulated FS circuit: %d constraints", ccs.GetNbConstraints())

	// Step 4: Check that the circuit is satisfied
	fullWitness, err := frontend.NewWitness(&witness, ecc.BN254.ScalarField())
	require.NoError(t, err)
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err, "GnarkFSKoalagnark should produce correct output on BN254")
}

// TestKoalagnarkCompileOnBN254 at minimum validates that the circuit compiles
// on BN254 without errors (even if IsSolved might fail due to hasher differences).
func TestKoalagnarkCompileOnBN254(t *testing.T) {
	var circuit KoalagnarkBN254Circuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, &circuit)
	require.NoError(t, err, "GnarkFSKoalagnark circuit should compile on BN254")

	// Cast to check it's a valid constraint system
	require.NotNil(t, ccs)
	require.True(t, ccs.GetNbConstraints() > 0, "circuit should have constraints")

	switch ccs := ccs.(type) {
	case constraint.ConstraintSystem:
		t.Logf("BN254 circuit compiled: %d constraints, field=%s",
			ccs.GetNbConstraints(), ccs.Field().String())
	}
}
