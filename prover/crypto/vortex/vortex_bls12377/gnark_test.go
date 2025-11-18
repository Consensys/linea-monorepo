//go:build !fuzzlight

package vortex

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	gnarkVortex "github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_bls12377"
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

var rng = rand.New(utils.NewRandSource(0))

// ------------------------------------------------------------
// test Encode and decode: BLS <--> Koalabear
func TestEncodeAndDecode(t *testing.T) {
	// test EncodeBLS12377ToKoalabear and DecodeKoalabearToBLS12377

	var elements fr.Element
	elements.SetRandom()
	bytes := elements.Bytes()
	elementsKoalaBear := EncodeBLS12377ToKoalabear(bytes)
	decodedElements := DecodeKoalabearToBLS12377(elementsKoalaBear)

	assert.Equal(t, types.AsBytes32(bytes[:]), decodedElements)

}

// ------------------------------------------------------------
// test Encode and Hash
type EncodeAndHashTestCircuit struct {
	Values [8]zk.WrappedVariable
	Result frontend.Variable
	Params GParams
}

func (c *EncodeAndHashTestCircuit) Define(api frontend.API) error {
	// Create constraints for encoding

	hasher, _ := c.Params.NoSisHasher(api)
	hasher.Reset()
	hashinput := EncodeWVsToFVs(api, c.Values[:])
	hasher.Write(hashinput...)
	digest := hasher.Sum()
	api.AssertIsEqual(digest, c.Result)
	return nil
}
func gnarkEncodeAndHashCircuitWitness() (*EncodeAndHashTestCircuit, *EncodeAndHashTestCircuit) {
	var intValues [8]field.Element
	var values [8]zk.WrappedVariable

	for i := 0; i < 8; i++ {
		intValues[i] = field.PseudoRand(rng)
		values[i] = zk.ValueOf(intValues[i].String())
	}

	intBytes := EncodeKoalabearsToBytes(intValues[:])
	hasher := smt_bls12377.Poseidon2()
	hasher.Write(intBytes)
	var expectedResult fr.Element
	expectedResult.SetBytes(hasher.Sum(nil))

	// Create test instance
	var circuit EncodeAndHashTestCircuit
	circuit.Values = values
	circuit.Result = expectedResult.String()
	circuit.Params.HasherFunc = makePoseidon2Hasherfunc
	circuit.Params.NoSisHasher = makePoseidon2Hasherfunc

	// Create a witness with test values
	var witness EncodeAndHashTestCircuit
	witness.Values = values
	witness.Result = expectedResult.String()

	return &circuit, &witness

}
func TestEncodeAndHashCircuit(t *testing.T) {

	{
		circuit, witness := gnarkEncodeAndHashCircuitWitness()

		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

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

func gnarkComputeLagrangeCircuitWitness(s int) (*ComputeLagrangeCircuit, *ComputeLagrangeCircuit) {
	d := fft.NewDomain(uint64(s))
	zeta := fext.PseudoRand(rng)

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

	return &circuit, &witness
}

func TestComputeLagrangeCircuit(t *testing.T) {

	s := 16

	{
		circuit, witness := gnarkComputeLagrangeCircuitWitness(s)

		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}
	{
		circuit, witness := gnarkComputeLagrangeCircuitWitness(s)

		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

}

// ------------------------------------------------------------
// test FFT inverse Ext: FFTInverseExt

type FFTInverseExtCircuit struct {
	Domain fft.Domain
	P      []gnarkfext.E4Gen
	R      []gnarkfext.E4Gen
}

func (circuit *FFTInverseExtCircuit) Define(api frontend.API) error {
	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		return err
	}
	f, err := FFTInverseExt(api, circuit.P, circuit.Domain.GeneratorInv, circuit.Domain.Cardinality)
	if err != nil {
		return err
	}

	for i := 0; i < len(f); i++ {
		ext4.AssertIsEqual(&circuit.R[i], &f[i])
	}

	return nil
}

func gnarkFFTInverseExtCircuitWitness(s int) (*FFTInverseExtCircuit, *FFTInverseExtCircuit) {
	d := fft.NewDomain(uint64(s))

	// prepare witness
	p := make([]fext.Element, s)
	for i := 0; i < s; i++ {
		p[i] = fext.PseudoRand(rng)
	}
	r := make([]fext.Element, s)
	copy(r, p)
	d.FFTInverseExt(r, fft.DIF)
	fft.BitReverse(r)
	var witness FFTInverseExtCircuit
	witness.P = make([]gnarkfext.E4Gen, s)
	witness.R = make([]gnarkfext.E4Gen, s)

	for i := 0; i < s; i++ {
		witness.P[i] = gnarkfext.NewE4Gen(p[i])
		witness.R[i] = gnarkfext.NewE4Gen(r[i])
	}

	var circuit FFTInverseExtCircuit
	circuit.P = make([]gnarkfext.E4Gen, s)
	circuit.R = make([]gnarkfext.E4Gen, s)
	circuit.Domain = *d
	return &circuit, &witness

}

func TestFFTInverseExtCircuit(t *testing.T) {

	s := 16

	{
		circuit, witness := gnarkFFTInverseExtCircuitWitness(s)

		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

	{
		circuit, witness := gnarkFFTInverseExtCircuitWitness(s)

		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

}

// ------------------------------------------------------------
// test AssertIsCodeWordExt

type AssertIsCodeWordExtCircuit struct {
	rate uint64
	d    *fft.Domain
	P    []gnarkfext.E4Gen `gnark:",public"`
}

func (circuit *AssertIsCodeWordExtCircuit) Define(api frontend.API) error {
	return assertIsCodeWordExt(api, circuit.P, circuit.d.GeneratorInv, circuit.d.Cardinality, circuit.rate)

}
func gnarkAssertIsCodeWordExtCircuitWitness(size int) (*AssertIsCodeWordExtCircuit, *AssertIsCodeWordExtCircuit) {
	d := fft.NewDomain(uint64(size))
	rate := 2
	p := make([]fext.Element, size)
	for i := 0; i < (size / rate); i++ {
		p[i] = fext.PseudoRand(rng)
	}
	d.FFTExt(p, fft.DIF)
	fft.BitReverse(p)

	var witness AssertIsCodeWordExtCircuit
	witness.P = make([]gnarkfext.E4Gen, size)
	witness.rate = uint64(rate)
	witness.d = d
	for i := 0; i < size; i++ {
		witness.P[i] = gnarkfext.NewE4Gen(p[i])
	}

	// compile the circuit
	var circuit AssertIsCodeWordExtCircuit
	circuit.P = make([]gnarkfext.E4Gen, size)
	circuit.d = d
	circuit.rate = uint64(rate)
	return &circuit, &witness
}

func TestAssertIsCodeWordExt(t *testing.T) {

	// generate witness
	size := 4

	{
		circuit, witness := gnarkAssertIsCodeWordExtCircuitWitness(size)

		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

	{
		circuit, witness := gnarkAssertIsCodeWordExtCircuitWitness(size)

		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}
}

// ------------------------------------------------------------
// test EvaluateLagrange: gnarkEvaluateLagrangeExt

type EvaluateLagrangeCircuit struct {
	P []gnarkfext.E4Gen
	X gnarkfext.E4Gen `gnark:",public"`
	Y gnarkfext.E4Gen
	d *fft.Domain
}

func (circuit *EvaluateLagrangeCircuit) Define(api frontend.API) error {

	res := gnarkEvaluateLagrangeExt(api, circuit.P, circuit.X, circuit.d.Generator, circuit.d.Cardinality)
	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		return err
	}
	ext4.AssertIsEqual(&res, &circuit.Y)

	return nil
}

func getEvaluateLagrangeExtCircuitWitness(size int) (*EvaluateLagrangeCircuit, *EvaluateLagrangeCircuit) {

	pCan := make([]fext.Element, size)
	for i := 0; i < size; i++ {
		pCan[i] = fext.PseudoRand(rng)
	}
	d := fft.NewDomain(uint64(size))

	x := fext.PseudoRand(rng)
	var y fext.Element
	for i := 0; i < size; i++ {
		y.Mul(&y, &x)
		y.Add(&y, &pCan[size-1-i])
	}

	d.FFTExt(pCan, fft.DIF)
	fft.BitReverse(pCan)

	var circuit, witness EvaluateLagrangeCircuit
	circuit.P = make([]gnarkfext.E4Gen, size)
	circuit.d = d

	witness.P = make([]gnarkfext.E4Gen, size)
	witness.X = gnarkfext.NewE4Gen(x)
	witness.Y = gnarkfext.NewE4Gen(y)
	for i := 0; i < size; i++ {
		witness.P[i] = gnarkfext.NewE4Gen(pCan[i])
	}

	return &circuit, &witness
}

func TestEvaluateLagrangeExtCircuit(t *testing.T) {

	size := 64

	{
		circuit, witness := getEvaluateLagrangeExtCircuitWitness(size)

		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

	{
		circuit, witness := getEvaluateLagrangeExtCircuitWitness(size)

		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, ecc.BLS12_377.ScalarField())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

}

// ------------------------------------------------------------
// test linear combination : gnarkEvalCanonical

type LinearCombinationCircuit struct {
	P          []zk.WrappedVariable
	RandomCoin gnarkfext.E4Gen `gnark:",public"`
	Y          gnarkfext.E4Gen
}

func (circuit *LinearCombinationCircuit) Define(api frontend.API) error {

	res := gnarkEvalCanonical(api, circuit.P, circuit.RandomCoin)
	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		return err
	}
	ext4.AssertIsEqual(&res, &circuit.Y)

	return nil
}

func gnarkEvalCanonicalCircuitWitness(size int) (*LinearCombinationCircuit, *LinearCombinationCircuit) {

	pCan := make([]field.Element, size)
	for i := 0; i < size; i++ {
		pCan[i] = field.PseudoRand(rng)
	}
	randomCoin := fext.PseudoRand(rng)
	y := gnarkVortex.EvalBasePolyHorner(pCan, randomCoin)

	var circuit, witness LinearCombinationCircuit
	circuit.P = make([]zk.WrappedVariable, size)

	witness.P = make([]zk.WrappedVariable, size)
	witness.RandomCoin = gnarkfext.NewE4Gen(randomCoin)
	witness.Y = gnarkfext.NewE4Gen(y)
	for i := 0; i < size; i++ {
		witness.P[i] = zk.ValueOf(pCan[i].String())
	}

	return &circuit, &witness
}

func TestLinearCombinationCircuit(t *testing.T) {

	size := 64

	{
		circuit, witness := gnarkEvalCanonicalCircuitWitness(size)

		ccs, err := frontend.CompileU32(koalabear.Modulus(), scs.NewBuilder, circuit)
		assert.NoError(t, err)

		fullWitness, err := frontend.NewWitness(witness, koalabear.Modulus())
		assert.NoError(t, err)
		err = ccs.IsSolved(fullWitness)
		assert.NoError(t, err)
	}

	{
		circuit, witness := getEvaluateLagrangeExtCircuitWitness(size)

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
	proof *vortex.OpeningProof,
	randomCoin fext.Element,
	x fext.Element,
	yLists [][]fext.Element,
	entryList []int,
	roots []types.Bytes32,
	merkleProofs [][]smt_bls12377.Proof,
) {
	x = fext.RandomElement()
	randomCoin = fext.RandomElement()
	entryList = []int{1, 5, 19, 645}
	blsParams := NewParams(blowUpFactor, polySize, nPolys*nCommitments, ringsis.StdParams, smt_bls12377.Poseidon2, smt_bls12377.Poseidon2)
	koalabearParams := vortex.NewParams(blowUpFactor, polySize, nPolys*nCommitments, ringsis.StdParams, poseidon2_koalabear.NewMDHasher, poseidon2_koalabear.NewMDHasher)

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
	proof = &vortex.OpeningProof{}
	proof.LinearCombination = koalabearParams.InitOpeningWithLC(utils.Join(polyLists...), randomCoin)

	merkleProofs = proof.GnarkComplete(entryList, committedMatrices, trees)

	// Check the proof
	err := VerifyOpening(&VerifierInputs{
		AlgebraicCheckInputs: vortex.AlgebraicCheckInputs{
			Koalabear_Params: *koalabearParams,
			X:                x,
			Ys:               yLists,
			OpeningProof:     vortex.OpeningProof(*proof),
			RandomCoin:       randomCoin,
			EntryList:        entryList,
		},
		BLS12_377_Params:         *blsParams,
		MerkleProofs:             merkleProofs,
		MerkleRoots:              roots,
		IsSISReplacedByPoseidon2: isSISReplacedByPoseidon2,
	})

	require.NoError(t, err)

	return proof, randomCoin, x, yLists, entryList, roots, merkleProofs
}

func gnarkVerifyOpeningCircuitMerkleTreeCircuitWitness(t *testing.T) (*VerifyOpeningCircuitMerkleTree, *VerifyOpeningCircuitMerkleTree) {

	// generate witness
	nCommitments := 4
	nPolys := 15
	polySize := 1 << 10
	blowUpFactor := 2
	proof, randomCoin, x, ys, entryList, roots, merkleProofs := getProofVortexNCommitmentsWithMerkleNoSis(t, nCommitments, nPolys, polySize, blowUpFactor)

	rsSize := blowUpFactor * polySize
	rsDomainSize := uint64(rsSize)
	rsDomain := fft.NewDomain(rsDomainSize)
	var witness VerifyOpeningCircuitMerkleTree
	witness.Proof.RsDomain = rsDomain
	witness.Proof.Rate = uint64(blowUpFactor)
	AllocateCircuitVariablesWithMerkleTree(&witness, *proof, ys, entryList, roots, merkleProofs)
	AssignCicuitVariablesWithMerkleTree(&witness, *proof, ys, entryList, roots, merkleProofs)
	witness.RandomCoin = gnarkfext.NewE4Gen(randomCoin)
	witness.X = gnarkfext.NewE4Gen(x)
	witness.Params.HasherFunc = makePoseidon2Hasherfunc
	witness.Params.NoSisHasher = makePoseidon2Hasherfunc

	// compile the circuit
	var circuit VerifyOpeningCircuitMerkleTree
	circuit.Proof.LinearCombination = make([]gnarkfext.E4Gen, rsSize)
	circuit.Proof.Rate = uint64(blowUpFactor)
	circuit.Proof.RsDomain = rsDomain
	circuit.Params.HasherFunc = makePoseidon2Hasherfunc
	circuit.Params.NoSisHasher = makePoseidon2Hasherfunc

	AllocateCircuitVariablesWithMerkleTree(&circuit, *proof, ys, entryList, roots, merkleProofs)

	return &circuit, &witness
}

func TestGnarkVortexNCommitmentsWithMerkleNoSis(t *testing.T) {
	{
		circuit, witness := gnarkVerifyOpeningCircuitMerkleTreeCircuitWitness(t)

		ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circuit, frontend.IgnoreUnconstrainedInputs())
		if err != nil {
			t.Fatal(err)
		}

		// solve the circuit
		twitness, err := frontend.NewWitness(witness, ecc.BLS12_377.ScalarField())
		if err != nil {
			t.Fatal(err)
		}
		err = ccs.IsSolved(twitness)
		if err != nil {
			t.Fatal(err)
		}
	}

}
func makePoseidon2Hasherfunc(api frontend.API) (poseidon2_bls12377.GnarkMDHasher, error) {

	h, err := poseidon2_bls12377.NewGnarkMDHasher(api)

	return h, err
}
