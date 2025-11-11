//go:build !fuzzlight

package vortex

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/std/hash"
	gposeidon2 "github.com/consensys/gnark/std/hash/poseidon2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_bls12377"
	"github.com/consensys/linea-monorepo/prover/crypto/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

// ------------------------------------------------------------
// test EncodeWVsToFV and Encode8KoalabearToBigInt
type EncodeTestCircuit struct {
	Values [8]zk.WrappedVariable
	Result frontend.Variable
}

func (c *EncodeTestCircuit) Define(api frontend.API) error {
	// Create constraints for encoding
	result := EncodeWVsToFV(api, c.Values)
	api.AssertIsEqual(result, c.Result)
	return nil
}

func TestEncodeCircuit(t *testing.T) {
	var intValues [8]field.Element
	var values [8]zk.WrappedVariable

	for i := 0; i < 8; i++ {
		intValues[i].SetRandom()
		values[i] = zk.ValueOf(intValues[i].String())
	}

	// Calculate expected result manually using big.Int (for validation)
	expectedResult := Encode8KoalabearToBigInt(intValues)

	// Create test instance
	var circuit EncodeTestCircuit
	circuit.Values = values
	circuit.Result = expectedResult.String()

	// Compile the circuit for BLS12-377
	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit)

	// Assert compilation worked
	assert.NoError(t, err)

	// Create a witness with test values
	var witness EncodeTestCircuit
	witness.Values = values
	witness.Result = expectedResult.String()

	// Convert witness to frontend-compatible witness
	fullWitness, err := frontend.NewWitness(&witness, ecc.BLS12_377.ScalarField())
	assert.NoError(t, err)

	// Verify the circuit satisfies constraints with the witness
	err = ccs.IsSolved(fullWitness)
	assert.NoError(t, err)

}

// ------------------------------------------------------------
// test computeLagrange: gnarkComputeLagrangeAtZ

func evalPoly(p []field.Element, z fext.Element) fext.Element {
	var res fext.Element
	for i := len(p) - 1; i >= 0; i-- {
		res.Mul(&res, &z)
		fext.AddByBase(&res, &res, &p[i])
	}
	return res
}

type ComputeLagrangeCircuit struct {
	Domain fft.Domain
	Zeta   gnarkfext.E4Gen   `gnark:",public"` // random variable
	Li     []gnarkfext.E4Gen // expected results
}

func (circuit *ComputeLagrangeCircuit) Define(api frontend.API) error {

	n := circuit.Domain.Cardinality
	gen := circuit.Domain.Generator
	r := gnarkComputeLagrangeAtZ(api, circuit.Zeta, gen, n)

	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		return err
	}

	for i := 0; i < len(r); i++ {
		ext4.AssertIsEqual(&r[i], &circuit.Li[i])
	}

	return nil
}

func TestComputeLagrangeCircuit(t *testing.T) {

	s := 16
	d := fft.NewDomain(uint64(s))
	var zeta fext.Element
	zeta.SetRandom()

	// prepare witness
	var witness ComputeLagrangeCircuit
	witness.Zeta = gnarkfext.NewE4Gen(zeta)
	witness.Li = make([]gnarkfext.E4Gen, s)
	for i := 0; i < s; i++ {
		buf := make([]field.Element, s)
		buf[i].SetOne()
		d.FFTInverse(buf, fft.DIF)
		fft.BitReverse(buf)
		li := evalPoly(buf, zeta)
		witness.Li[i] = gnarkfext.NewE4Gen(li)
	}

	var circuit ComputeLagrangeCircuit
	circuit.Domain = *d
	circuit.Li = make([]gnarkfext.E4Gen, s)

	// compile...
	builder := scs.NewBuilder[constraint.U32]
	ccs, err := frontend.CompileGeneric[constraint.U32](field.Modulus(), builder, &circuit)
	if err != nil {
		t.Fatal(err)
	}

	// solve the circuit
	twitness, err := frontend.NewWitness(&witness, field.Modulus())
	if err != nil {
		t.Fatal(err)
	}
	err = ccs.IsSolved(twitness)
	if err != nil {
		t.Fatal(err)
	}

}

// ------------------------------------------------------------
// test FFT inverse: FFTInverse

type FFTInverseCircuit struct {
	Domain fft.Domain
	P      []zk.WrappedVariable
	R      []zk.WrappedVariable
}

func (circuit *FFTInverseCircuit) Define(api frontend.API) error {
	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		return err
	}
	f, err := FFTInverse(api, circuit.P, circuit.Domain.GeneratorInv, circuit.Domain.Cardinality)
	if err != nil {
		return err
	}

	for i := 0; i < len(f); i++ {
		apiGen.AssertIsEqual(circuit.R[i], f[i])
	}

	return nil
}

func TestFFTInverseCircuit(t *testing.T) {

	s := 16
	d := fft.NewDomain(uint64(s))

	// prepare witness
	p := make([]field.Element, s)
	for i := 0; i < s; i++ {
		p[i].SetRandom()
	}
	r := make([]field.Element, s)
	copy(r, p)
	d.FFTInverse(r, fft.DIF)
	fft.BitReverse(r)
	var witness FFTInverseCircuit
	witness.P = make([]zk.WrappedVariable, s)
	witness.R = make([]zk.WrappedVariable, s)

	for i := 0; i < s; i++ {
		witness.P[i] = zk.ValueOf(p[i].String())
		witness.R[i] = zk.ValueOf(r[i].String())
	}

	var circuit FFTInverseCircuit
	circuit.P = make([]zk.WrappedVariable, s)
	circuit.R = make([]zk.WrappedVariable, s)
	circuit.Domain = *d

	// compile...
	builder := scs.NewBuilder[constraint.U32]
	ccs, err := frontend.CompileGeneric[constraint.U32](field.Modulus(), builder, &circuit)
	if err != nil {
		t.Fatal(err)
	}

	// solve the circuit
	twitness, err := frontend.NewWitness(&witness, field.Modulus())
	if err != nil {
		t.Fatal(err)
	}
	err = ccs.IsSolved(twitness)
	if err != nil {
		t.Fatal(err)
	}

}

// ------------------------------------------------------------
// test AssertIsCodeWord

// type AssertIsCodeWordCircuit struct {
// 	rate uint64
// 	d    *fft.Domain
// 	P    []zk.WrappedVariable `gnark:",public"`
// }

// func (circuit *AssertIsCodeWordCircuit) Define(api frontend.API) error {
// 	return assertIsCodeWord(api, circuit.P, circuit.d.GeneratorInv, circuit.d.Cardinality, circuit.rate)

// }

/*
func TestAssertIsCodeWord(t *testing.T) {

	// generate witness
	size := 2048
	d := fft.NewDomain(uint64(size))
	rate := 2
	p := make([]field.Element, size)
	for i := 0; i < (size / rate); i++ {
		p[i].SetRandom()
	}
	d.FFT(p, fft.DIF)
	fft.BitReverse(p)

	var witness AssertIsCodeWordCircuit
	witness.P = make([]zk.WrappedVariable, size)
	witness.rate = uint64(rate)
	witness.d = d
	for i := 0; i < size; i++ {
		witness.P[i] = p[i].String()
	}

	// compile the circuit
	var circuit AssertIsCodeWordCircuit
	circuit.P = make([]zk.WrappedVariable, size)
	circuit.d = d
	circuit.rate = uint64(rate)
	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit)
	if err != nil {
		t.Fatal(err)
	}

	// solve the circuit
	twitness, err := frontend.NewWitness(&witness, ecc.BLS12_377.ScalarField())
	if err != nil {
		t.Fatal(err)
	}
	err = ccs.IsSolved(twitness)
	if err != nil {
		t.Fatal(err)
	}

}
*/

// ------------------------------------------------------------
// test EvaluateLagrange: gnarkEvaluateLagrange

type EvaluateLagrangeCircuit struct {
	P []zk.WrappedVariable
	X gnarkfext.E4Gen `gnark:",public"`
	Y gnarkfext.E4Gen
	d *fft.Domain
}

func (circuit *EvaluateLagrangeCircuit) Define(api frontend.API) error {

	res := gnarkEvaluateLagrange(api, circuit.P, circuit.X, circuit.d.Generator, circuit.d.Cardinality)
	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		return err
	}
	ext4.AssertIsEqual(&res, &circuit.Y)

	return nil
}

func getEvaluateLagrangeCircuitWitness(size int) (*EvaluateLagrangeCircuit, *EvaluateLagrangeCircuit) {

	pCan := make([]field.Element, size)
	for i := 0; i < size; i++ {
		pCan[i].SetRandom()
	}
	d := fft.NewDomain(uint64(size))

	var x fext.Element
	x.SetRandom()
	var y fext.Element
	for i := 0; i < size; i++ {
		y.Mul(&y, &x)
		fext.AddByBase(&y, &y, &pCan[size-1-i])
	}

	d.FFT(pCan, fft.DIF)
	fft.BitReverse(pCan)

	var circuit, witness EvaluateLagrangeCircuit
	circuit.P = make([]zk.WrappedVariable, size)
	circuit.d = d

	witness.P = make([]zk.WrappedVariable, size)
	witness.X = gnarkfext.NewE4Gen(x)
	witness.Y = gnarkfext.NewE4Gen(y)
	for i := 0; i < size; i++ {
		witness.P[i] = zk.ValueOf(pCan[i].String())
	}

	return &circuit, &witness
}

func TestEvaluateLagrangeCircuit(t *testing.T) {

	size := 64

	{
		circuit, witness := getEvaluateLagrangeCircuitWitness(size)

		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

	{
		circuit, witness := getEvaluateLagrangeCircuitWitness(size)

		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

}

// ------------------------------------------------------------
// N Commitments with Merkle opening

func getProofVortexNCommitmentsWithMerkleNoSis(t *testing.T, nCommitments, nPolys, polySize, blowUpFactor int) (
	proof *OpeningProof,
	randomCoin fext.Element,
	x fext.Element,
	yLists [][]fext.Element,
	entryList []int,
	roots []types.Bytes32,
) {

	x = fext.RandomElement()
	randomCoin = fext.RandomElement()
	entryList = []int{1, 5, 19, 645}

	blsParams := NewParams(blowUpFactor, polySize, nPolys*nCommitments, ringsis.StdParams, smt_bls12377.Poseidon2, smt_bls12377.Poseidon2)
	koalabearParams := vortex.NewParams(blowUpFactor, polySize, nPolys*nCommitments, ringsis.StdParams, poseidon2_koalabear.Poseidon2, poseidon2_koalabear.Poseidon2)

	polyLists := make([][]smartvectors.SmartVector, nCommitments)
	yLists = make([][]fext.Element, nCommitments)
	for j := range polyLists {
		// Polynomials to commit to
		polys := make([]smartvectors.SmartVector, nPolys)
		ys := make([]fext.Element, nPolys)
		for i := range polys {
			polys[i] = smartvectors.Rand(polySize)
			ys[i] = smartvectors.EvaluateBasePolyLagrange(polys[i], x)
		}
		polyLists[j] = polys
		yLists[j] = ys
	}

	// Commits to it
	roots = make([]types.Bytes32, nCommitments)
	trees := make([]*smt_bls12377.Tree, nCommitments)
	committedMatrices := make([]vortex.EncodedMatrix, nCommitments)
	isSISReplacedByPoseidon2 := make([]bool, nCommitments)
	for j := range trees {
		// encode before committing
		if len(polyLists[j]) > koalabearParams.MaxNbRows {
			utils.Panic("too many rows: %v, capacity is %v\n", len(polyLists[j]), koalabearParams.MaxNbRows)
		}

		profiling.TimeIt(func() {
			committedMatrices[j] = koalabearParams.EncodeRows(polyLists[j])
		})

		// As Gnark does not support SIS, we commit without SIS hashing
		trees[j], _ = blsParams.CommitMerkleWithoutSIS(committedMatrices[j])
		roots[j] = trees[j].Root
		// We set the SIS replaced by Poseidon2 to true, as Gnark does not support SIS
		isSISReplacedByPoseidon2[j] = true
	}

	// Generate the proof
	proof.LinearCombination = koalabearParams.InitOpeningWithLC(utils.Join(polyLists...), randomCoin)
	proof.Complete(entryList, committedMatrices, trees)

	// Check the proof
	err := VerifyOpening(&VerifierInputs{
		Koalabear_Params:         *koalabearParams,
		BLS12_377_Params:         *blsParams,
		MerkleRoots:              roots,
		X:                        x,
		Ys:                       yLists,
		OpeningProof:             *proof,
		RandomCoin:               randomCoin,
		EntryList:                entryList,
		IsSISReplacedByPoseidon2: isSISReplacedByPoseidon2,
	})
	require.NoError(t, err)

	return proof, randomCoin, x, yLists, entryList, roots
}

func TestGnarkVortexNCommitmentsWithMerkleNoSis(t *testing.T) {

	// generate witness
	nCommitments := 4
	nPolys := 15
	polySize := 1 << 10
	blowUpFactor := 2
	proof, randomCoin, x, ys, entryList, roots := getProofVortexNCommitmentsWithMerkleNoSis(t, nCommitments, nPolys, polySize, blowUpFactor)

	rsSize := blowUpFactor * polySize
	rsDomainSize := uint64(rsSize)
	rsDomain := fft.NewDomain(rsDomainSize)
	var witness VerifyOpeningCircuitMerkleTree
	witness.Proof.RsDomain = rsDomain
	witness.Proof.Rate = uint64(blowUpFactor)
	AllocateCircuitVariablesWithMerkleTree(&witness, *proof, ys, entryList, roots)
	AssignCicuitVariablesWithMerkleTree(&witness, *proof, ys, entryList, roots)
	witness.RandomCoin = gnarkfext.NewE4Gen(randomCoin)
	witness.X = gnarkfext.NewE4Gen(x)
	witness.Params.HasherFunc = makePoseidon2Hasherfunc
	witness.Params.NoSisHasher = makePoseidon2Hasherfunc

	// compile the circuit
	var circuit VerifyOpeningCircuitMerkleTree
	circuit.Proof.LinearCombination = make([]zk.WrappedVariable, rsSize)
	circuit.Proof.Rate = uint64(blowUpFactor)
	circuit.Proof.RsDomain = rsDomain
	circuit.Params.HasherFunc = makePoseidon2Hasherfunc
	circuit.Params.NoSisHasher = makePoseidon2Hasherfunc

	AllocateCircuitVariablesWithMerkleTree(&circuit, *proof, ys, entryList, roots)
	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit, frontend.IgnoreUnconstrainedInputs())
	if err != nil {
		t.Fatal(err)
	}

	// solve the circuit
	twitness, err := frontend.NewWitness(&witness, ecc.BLS12_377.ScalarField())
	if err != nil {
		t.Fatal(err)
	}
	err = ccs.IsSolved(twitness)
	if err != nil {
		t.Fatal(err)
	}
}
func makePoseidon2Hasherfunc(api frontend.API) (hash.FieldHasher, error) {

	h, err := gposeidon2.NewMerkleDamgardHasher(api)

	return h, err
}
