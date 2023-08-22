//go:build !fuzzlight

package vortex2

import (
	// "github.com/consensys/accelerated-crypto-monorepo/crypto/ringsis"
	// "github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	// "github.com/consensys/accelerated-crypto-monorepo/maths/field"

	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/mimc"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/ringsis"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/smt"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/std/hash"
	gmimc "github.com/consensys/gnark/std/hash/mimc"
	"github.com/stretchr/testify/require"
)

// ------------------------------------------------------------
// test computeLagrange

func evalPoly(p []fr.Element, z fr.Element) fr.Element {
	var res fr.Element
	n := len(p)
	for i := 0; i < len(p); i++ {
		res.Mul(&res, &z)
		res.Add(&res, &p[n-1-i])
	}
	return res
}

type ComputeLagrangeCircuit struct {
	Domain fft.Domain
	Zeta   frontend.Variable   `gnark:"public"` // random variable
	Li     []frontend.Variable // expected results
}

func (circuit *ComputeLagrangeCircuit) Define(api frontend.API) error {

	n := circuit.Domain.Cardinality
	gen := circuit.Domain.Generator
	r := gnarkComputeLagrangeAtZ(api, circuit.Zeta, gen, n)

	for i := 0; i < len(r); i++ {
		api.AssertIsEqual(r[i], circuit.Li[i])
	}

	return nil
}

func TestComputeLagrangeCircuit(t *testing.T) {

	s := 16
	d := fft.NewDomain(uint64(s))
	var zeta fr.Element
	zeta.SetRandom()

	// prepare witness
	var witness ComputeLagrangeCircuit
	witness.Zeta = zeta.String()
	witness.Li = make([]frontend.Variable, s)
	for i := 0; i < s; i++ {
		buf := make([]fr.Element, s)
		buf[i].SetOne()
		d.FFTInverse(buf, fft.DIF)
		fft.BitReverse(buf)
		li := evalPoly(buf, zeta)
		witness.Li[i] = li.String()
	}

	var circuit ComputeLagrangeCircuit
	circuit.Domain = *d
	circuit.Li = make([]frontend.Variable, s)

	// compile...
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, &circuit)
	if err != nil {
		t.Fatal(err)
	}

	// solve the circuit
	twitness, err := frontend.NewWitness(&witness, ecc.BN254.ScalarField())
	if err != nil {
		t.Fatal(err)
	}
	err = ccs.IsSolved(twitness)
	if err != nil {
		t.Fatal(err)
	}

}

// ------------------------------------------------------------
// test FFT inverse

type FFTInverseCircuit struct {
	Domain fft.Domain
	P      []frontend.Variable
	R      []frontend.Variable
}

func (circuit *FFTInverseCircuit) Define(api frontend.API) error {

	f, err := FFTInverse(api, circuit.P, circuit.Domain.GeneratorInv, circuit.Domain.Cardinality)
	if err != nil {
		return err
	}

	for i := 0; i < len(f); i++ {
		api.AssertIsEqual(circuit.R[i], f[i])
	}

	return nil
}

func TestFFTInverseCircuit(t *testing.T) {

	s := 16
	d := fft.NewDomain(uint64(s))

	// prepare witness
	p := make([]fr.Element, s)
	for i := 0; i < s; i++ {
		p[i].SetRandom()
	}
	r := make([]fr.Element, s)
	copy(r, p)
	d.FFTInverse(r, fft.DIF)
	fft.BitReverse(r)
	var witness FFTInverseCircuit
	witness.P = make([]frontend.Variable, s)
	witness.R = make([]frontend.Variable, s)

	for i := 0; i < s; i++ {
		witness.P[i] = p[i].String()
		witness.R[i] = r[i].String()
	}

	var circuit FFTInverseCircuit
	circuit.P = make([]frontend.Variable, s)
	circuit.R = make([]frontend.Variable, s)
	circuit.Domain = *d

	// compile...
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, &circuit)
	if err != nil {
		t.Fatal(err)
	}

	// solve the circuit
	twitness, err := frontend.NewWitness(&witness, ecc.BN254.ScalarField())
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

type AssertIsCodeWordCircuit struct {
	rate uint64
	d    *fft.Domain
	P    []frontend.Variable `gnark:"public"`
}

func (circuit *AssertIsCodeWordCircuit) Define(api frontend.API) error {

	return assertIsCodeWord(api, circuit.P, circuit.d.GeneratorInv, circuit.d.Cardinality, circuit.rate)

}

func TestAssertIsCodeWord(t *testing.T) {

	// generate witness
	size := 2048
	d := fft.NewDomain(uint64(size))
	rate := 2
	p := make([]fr.Element, size)
	for i := 0; i < (size / rate); i++ {
		p[i].SetRandom()
	}
	d.FFT(p, fft.DIF)
	fft.BitReverse(p)

	var witness AssertIsCodeWordCircuit
	witness.P = make([]frontend.Variable, size)
	witness.rate = uint64(rate)
	witness.d = d
	for i := 0; i < size; i++ {
		witness.P[i] = p[i].String()
	}

	// compile the circuit
	var circuit AssertIsCodeWordCircuit
	circuit.P = make([]frontend.Variable, size)
	circuit.d = d
	circuit.rate = uint64(rate)
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, &circuit)
	if err != nil {
		t.Fatal(err)
	}

	// solve the circuit
	twitness, err := frontend.NewWitness(&witness, ecc.BN254.ScalarField())
	if err != nil {
		t.Fatal(err)
	}
	err = ccs.IsSolved(twitness)
	if err != nil {
		t.Fatal(err)
	}

}

// ------------------------------------------------------------
// Interpolate

type InterpolateCircuit struct {
	P []frontend.Variable
	X frontend.Variable `gnark:"public"`
	R frontend.Variable
	d *fft.Domain
}

func (circuit *InterpolateCircuit) Define(api frontend.API) error {

	res := gnarkInterpolate(api, circuit.P, circuit.X, circuit.d.Generator, circuit.d.Cardinality)

	api.AssertIsEqual(res, circuit.R)

	return nil
}

func TestInterpolateCircuit(t *testing.T) {

	// generate witness
	size := 16
	p := make([]fr.Element, size)
	for i := 0; i < size; i++ {
		p[i].SetRandom()
	}
	var x fr.Element
	x.SetRandom()
	var r fr.Element
	for i := 0; i < size; i++ {
		r.Mul(&r, &x)
		r.Add(&r, &p[len(p)-1-i])
	}

	d := fft.NewDomain(uint64(size))
	d.FFT(p, fft.DIF)
	fft.BitReverse(p)

	var witness InterpolateCircuit
	witness.P = make([]frontend.Variable, size)
	for i := 0; i < size; i++ {
		witness.P[i] = p[i].String()
	}
	witness.X = x
	witness.R = r

	// compile the circuit
	var circuit InterpolateCircuit
	circuit.P = make([]frontend.Variable, size)
	circuit.d = d
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, &circuit)
	if err != nil {
		t.Fatal(err)
	}

	// solve the circuit
	twitness, err := frontend.NewWitness(&witness, ecc.BN254.ScalarField())
	if err != nil {
		t.Fatal(err)
	}
	err = ccs.IsSolved(twitness)
	if err != nil {
		t.Fatal(err)
	}

}

// ------------------------------------------------------------
// One Commitment opening

func getProofVortexOneCommitment(t *testing.T, nPolys, polySize, blowUpFactor int) (
	proof *Proof,
	randomCoin field.Element,
	x field.Element,
	ys [][]field.Element,
	entryList []int,
	key ringsis.Key,
	commitments []Commitment,
) {

	x = field.NewElement(478)
	randomCoin = field.NewElement(1523)
	entryList = []int{1, 5, 19, 645}

	params := NewParams(blowUpFactor, polySize, nPolys, ringsis.StdParams)

	// Polynomials to commit to
	polys := make([]smartvectors.SmartVector, nPolys)
	_ys := make([]field.Element, nPolys)
	for i := range polys {
		polys[i] = smartvectors.Rand(polySize)
		_ys[i] = smartvectors.Interpolate(polys[i], x)
	}

	// Commits to it
	commitment, committedMatrix := params.Commit(polys)

	// Generate the proof
	proof = params.OpenWithLC(polys, randomCoin)
	proof.WithEntryList([]CommittedMatrix{committedMatrix}, entryList)

	// Check the proof
	ys = [][]field.Element{_ys}
	err := params.VerifyOpening([]Commitment{commitment}, proof, x, ys, randomCoin, entryList)
	require.NoError(t, err)

	key = params.Key
	commitments = []Commitment{commitment}
	return proof, randomCoin, x, ys, entryList, key, commitments

}

func TestGnarkVortexOneCommitment(t *testing.T) {

	// generate witness
	nPolys := 15
	polySize := 1 << 10
	blowUpFactor := 2
	proof, randomCoin, x, ys, entryList, key, commitments := getProofVortexOneCommitment(t, nPolys, polySize, blowUpFactor)

	rsSize := blowUpFactor * polySize
	rsDomainSize := uint64(rsSize)
	rsDomain := fft.NewDomain(rsDomainSize)
	var witness VerifyOpeningCircuit
	witness.Proof.RsDomain = rsDomain
	witness.Proof.Rate = uint64(blowUpFactor)
	AllocateCircuitVariables(&witness, *proof, ys, entryList, commitments)
	AssignCicuitVariables(&witness, *proof, ys, entryList, commitments)
	witness.RandomCoin = randomCoin.String()
	witness.X = x.String()
	witness.Params.Key = key

	// compile the circuit
	var circuit VerifyOpeningCircuit
	circuit.Proof.LinearCombination = make([]frontend.Variable, rsSize)
	circuit.Proof.Rate = uint64(blowUpFactor)
	circuit.Proof.RsDomain = rsDomain
	circuit.Params.Key = key
	circuit.Params.HasherFunc = makeMimcHasherfunc
	AllocateCircuitVariables(&circuit, *proof, ys, entryList, commitments)
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, &circuit, frontend.IgnoreUnconstrainedInputs())
	if err != nil {
		t.Fatal(err)
	}

	// solve the circuit
	twitness, err := frontend.NewWitness(&witness, ecc.BN254.ScalarField())
	if err != nil {
		t.Fatal(err)
	}
	err = ccs.IsSolved(twitness)
	if err != nil {
		t.Fatal(err)
	}
}

// ------------------------------------------------------------
// N Commitments opening
func getProofVortexNCommitments(t *testing.T, nCommitments, nPolys, polySize, blowUpFactor int) (
	proof *Proof,
	randomCoin field.Element,
	x field.Element,
	yLists [][]field.Element,
	entryList []int,
	key ringsis.Key,
	commitments []Commitment,
) {

	x = field.NewElement(478)
	randomCoin = field.NewElement(1523)
	entryList = []int{1, 5, 19, 645}

	params := NewParams(blowUpFactor, polySize, nPolys*nCommitments, ringsis.StdParams)

	polyLists := make([][]smartvectors.SmartVector, nCommitments)
	yLists = make([][]field.Element, nCommitments)
	for j := range polyLists {
		// Polynomials to commit to
		polys := make([]smartvectors.SmartVector, nPolys)
		ys := make([]field.Element, nPolys)
		for i := range polys {
			polys[i] = smartvectors.Rand(polySize)
			ys[i] = smartvectors.Interpolate(polys[i], x)
		}
		polyLists[j] = polys
		yLists[j] = ys
	}

	// Commits to it
	commitments = make([]Commitment, nCommitments)
	committedMatrices := make([]CommittedMatrix, nCommitments)
	for j := range commitments {
		commitments[j], committedMatrices[j] = params.Commit(polyLists[j])
	}

	// Generate the proof
	proof = params.OpenWithLC(utils.Join(polyLists...), randomCoin)
	proof.WithEntryList(committedMatrices, entryList)

	// Check the proof
	err := params.VerifyOpening(commitments, proof, x, yLists, randomCoin, entryList)
	require.NoError(t, err)

	key = params.Key
	return proof, randomCoin, x, yLists, entryList, key, commitments

}

func TestGnarkVortexNCommitments(t *testing.T) {

	// generate witness
	nCommitments := 4
	nPolys := 15
	polySize := 1 << 10
	blowUpFactor := 2
	proof, randomCoin, x, ys, entryList, key, commitments := getProofVortexNCommitments(t, nCommitments, nPolys, polySize, blowUpFactor)

	rsSize := blowUpFactor * polySize
	rsDomainSize := uint64(rsSize)
	rsDomain := fft.NewDomain(rsDomainSize)
	var witness VerifyOpeningCircuit
	witness.Proof.RsDomain = rsDomain
	witness.Proof.Rate = uint64(blowUpFactor)
	AllocateCircuitVariables(&witness, *proof, ys, entryList, commitments)
	AssignCicuitVariables(&witness, *proof, ys, entryList, commitments)
	witness.RandomCoin = randomCoin.String()
	witness.X = x.String()
	witness.Params.Key = key
	witness.Params.HasherFunc = makeMimcHasherfunc

	// compile the circuit
	var circuit VerifyOpeningCircuit
	circuit.Proof.LinearCombination = make([]frontend.Variable, rsSize)
	circuit.Proof.Rate = uint64(blowUpFactor)
	circuit.Proof.RsDomain = rsDomain
	circuit.Params.Key = key
	circuit.Params.HasherFunc = makeMimcHasherfunc
	AllocateCircuitVariables(&circuit, *proof, ys, entryList, commitments)
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, &circuit, frontend.IgnoreUnconstrainedInputs())
	if err != nil {
		t.Fatal(err)
	}

	// solve the circuit
	twitness, err := frontend.NewWitness(&witness, ecc.BN254.ScalarField())
	if err != nil {
		t.Fatal(err)
	}
	err = ccs.IsSolved(twitness)
	if err != nil {
		t.Fatal(err)
	}
}

// ------------------------------------------------------------
// N Commitments with Merkle opening

func getProofVortexNCommitmentsWithMerkle(t *testing.T, nCommitments, nPolys, polySize, blowUpFactor int) (
	proof *Proof,
	randomCoin field.Element,
	x field.Element,
	yLists [][]field.Element,
	entryList []int,
	key ringsis.Key,
	roots []hashtypes.Digest,
) {

	x = field.NewElement(478)
	randomCoin = field.NewElement(1523)
	entryList = []int{1, 5, 19, 645}

	params := NewParams(blowUpFactor, polySize, nPolys*nCommitments, ringsis.StdParams)
	params.WithMerkleMode(mimc.NewMiMC)

	polyLists := make([][]smartvectors.SmartVector, nCommitments)
	yLists = make([][]field.Element, nCommitments)
	for j := range polyLists {
		// Polynomials to commit to
		polys := make([]smartvectors.SmartVector, nPolys)
		ys := make([]field.Element, nPolys)
		for i := range polys {
			polys[i] = smartvectors.Rand(polySize)
			ys[i] = smartvectors.Interpolate(polys[i], x)
		}
		polyLists[j] = polys
		yLists[j] = ys
	}

	// Commits to it
	roots = make([]hashtypes.Digest, nCommitments)
	trees := make([]*smt.Tree, nCommitments)
	committedMatrices := make([]CommittedMatrix, nCommitments)
	for j := range trees {
		committedMatrices[j], trees[j], _ = params.CommitMerkle(polyLists[j])
		roots[j] = trees[j].Root
	}

	// Generate the proof
	proof = params.OpenWithLC(utils.Join(polyLists...), randomCoin)
	proof.WithEntryList(committedMatrices, entryList)
	proof.WithMerkleProof(trees, entryList)

	// Check the proof
	err := params.VerifyMerkle(roots, proof, x, yLists, randomCoin, entryList)
	require.NoError(t, err)

	key = params.Key
	return proof, randomCoin, x, yLists, entryList, key, roots

}

func TestGnarkVortexNCommitmentsWithMerkle(t *testing.T) {

	// generate witness
	nCommitments := 4
	nPolys := 15
	polySize := 1 << 10
	blowUpFactor := 2
	proof, randomCoin, x, ys, entryList, key, commitments := getProofVortexNCommitmentsWithMerkle(t, nCommitments, nPolys, polySize, blowUpFactor)

	rsSize := blowUpFactor * polySize
	rsDomainSize := uint64(rsSize)
	rsDomain := fft.NewDomain(rsDomainSize)
	var witness VerifyOpeningCircuitMerkleTree
	witness.Proof.RsDomain = rsDomain
	witness.Proof.Rate = uint64(blowUpFactor)
	AllocateCircuitVariablesWithMerkleTree(&witness, *proof, ys, entryList, commitments)
	AssignCicuitVariablesWithMerkleTree(&witness, *proof, ys, entryList, commitments)
	witness.RandomCoin = randomCoin.String()
	witness.X = x.String()
	witness.Params.Key = key

	// compile the circuit
	var circuit VerifyOpeningCircuitMerkleTree
	circuit.Proof.LinearCombination = make([]frontend.Variable, rsSize)
	circuit.Proof.Rate = uint64(blowUpFactor)
	circuit.Proof.RsDomain = rsDomain
	circuit.Params.Key = key
	circuit.Params.HasherFunc = makeMimcHasherfunc
	AllocateCircuitVariablesWithMerkleTree(&circuit, *proof, ys, entryList, commitments)
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, &circuit, frontend.IgnoreUnconstrainedInputs())
	if err != nil {
		t.Fatal(err)
	}

	// solve the circuit
	twitness, err := frontend.NewWitness(&witness, ecc.BN254.ScalarField())
	if err != nil {
		t.Fatal(err)
	}
	err = ccs.IsSolved(twitness)
	if err != nil {
		t.Fatal(err)
	}
}

func getProofVortexNCommitmentsWithMerkleNoSis(t *testing.T, nCommitments, nPolys, polySize, blowUpFactor int) (
	proof *Proof,
	randomCoin field.Element,
	x field.Element,
	yLists [][]field.Element,
	entryList []int,
	roots []hashtypes.Digest,
) {

	x = field.NewElement(478)
	randomCoin = field.NewElement(1523)
	entryList = []int{1, 5, 19, 645}

	params := NewParams(blowUpFactor, polySize, nPolys*nCommitments, ringsis.StdParams)
	params.WithMerkleMode(mimc.NewMiMC)
	params.RemoveSis(mimc.NewMiMC)

	polyLists := make([][]smartvectors.SmartVector, nCommitments)
	yLists = make([][]field.Element, nCommitments)
	for j := range polyLists {
		// Polynomials to commit to
		polys := make([]smartvectors.SmartVector, nPolys)
		ys := make([]field.Element, nPolys)
		for i := range polys {
			polys[i] = smartvectors.Rand(polySize)
			ys[i] = smartvectors.Interpolate(polys[i], x)
		}
		polyLists[j] = polys
		yLists[j] = ys
	}

	// Commits to it
	roots = make([]hashtypes.Digest, nCommitments)
	trees := make([]*smt.Tree, nCommitments)
	committedMatrices := make([]CommittedMatrix, nCommitments)
	for j := range trees {
		committedMatrices[j], trees[j], _ = params.CommitMerkle(polyLists[j])
		roots[j] = trees[j].Root
	}

	// Generate the proof
	proof = params.OpenWithLC(utils.Join(polyLists...), randomCoin)
	proof.WithEntryList(committedMatrices, entryList)
	proof.WithMerkleProof(trees, entryList)

	// Check the proof
	err := params.VerifyMerkle(roots, proof, x, yLists, randomCoin, entryList)
	require.NoError(t, err)

	return proof, randomCoin, x, yLists, entryList, roots
}

func TestGnarkVortexNCommitmentsWithMerkleNoSis(t *testing.T) {

	// generate witness
	nCommitments := 4
	nPolys := 15
	polySize := 1 << 10
	blowUpFactor := 2
	proof, randomCoin, x, ys, entryList, commitments := getProofVortexNCommitmentsWithMerkleNoSis(t, nCommitments, nPolys, polySize, blowUpFactor)

	rsSize := blowUpFactor * polySize
	rsDomainSize := uint64(rsSize)
	rsDomain := fft.NewDomain(rsDomainSize)
	var witness VerifyOpeningCircuitMerkleTree
	witness.Proof.RsDomain = rsDomain
	witness.Proof.Rate = uint64(blowUpFactor)
	AllocateCircuitVariablesWithMerkleTree(&witness, *proof, ys, entryList, commitments)
	AssignCicuitVariablesWithMerkleTree(&witness, *proof, ys, entryList, commitments)
	witness.RandomCoin = randomCoin.String()
	witness.X = x.String()
	witness.Params.HasherFunc = makeMimcHasherfunc
	witness.Params.NoSisHasher = makeMimcHasherfunc

	// compile the circuit
	var circuit VerifyOpeningCircuitMerkleTree
	circuit.Proof.LinearCombination = make([]frontend.Variable, rsSize)
	circuit.Proof.Rate = uint64(blowUpFactor)
	circuit.Proof.RsDomain = rsDomain
	circuit.Params.HasherFunc = makeMimcHasherfunc
	circuit.Params.NoSisHasher = makeMimcHasherfunc

	AllocateCircuitVariablesWithMerkleTree(&circuit, *proof, ys, entryList, commitments)
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, &circuit, frontend.IgnoreUnconstrainedInputs())
	if err != nil {
		t.Fatal(err)
	}

	// solve the circuit
	twitness, err := frontend.NewWitness(&witness, ecc.BN254.ScalarField())
	if err != nil {
		t.Fatal(err)
	}
	err = ccs.IsSolved(twitness)
	if err != nil {
		t.Fatal(err)
	}
}

func makeMimcHasherfunc(api frontend.API) (hash.FieldHasher, error) {
	h, err := gmimc.NewMiMC(api)
	return &h, err
}
