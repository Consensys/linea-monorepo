//go:build !fuzzlight

package vortex

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"

	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/std/hash"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/stretchr/testify/require"
)

// ------------------------------------------------------------
// test computeLagrange

func evalPoly(p []field.Element, z field.Element) field.Element {
	var res field.Element
	for i := len(p) - 1; i >= 0; i-- {
		res.Mul(&res, &z)
		res.Add(&res, &p[i])
	}
	return res
}

type ComputeLagrangeCircuit struct {
	Domain fft.Domain
	Zeta   frontend.Variable   `gnark:",public"` // random variable
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
	var zeta field.Element
	zeta.SetRandom()

	// prepare witness
	var witness ComputeLagrangeCircuit
	witness.Zeta = zeta.String()
	witness.Li = make([]frontend.Variable, s)
	for i := 0; i < s; i++ {
		buf := make([]field.Element, s)
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

// func TestFFTInverseCircuit(t *testing.T) {

// 	s := 16
// 	d := fft.NewDomain(uint64(s))

// 	// prepare witness
// 	p := make([]field.Element, s)
// 	for i := 0; i < s; i++ {
// 		p[i].SetRandom()
// 	}
// 	r := make([]field.Element, s)
// 	copy(r, p)
// 	d.FFTInverse(r, fft.DIF)
// 	fft.BitReverse(r)
// 	var witness FFTInverseCircuit
// 	witness.P = make([]frontend.Variable, s)
// 	witness.R = make([]frontend.Variable, s)

// 	for i := 0; i < s; i++ {
// 		witness.P[i] = p[i].String()
// 		witness.R[i] = r[i].String()
// 	}

// 	var circuit FFTInverseCircuit
// 	circuit.P = make([]frontend.Variable, s)
// 	circuit.R = make([]frontend.Variable, s)
// 	circuit.Domain = *d

// 	// compile...
// 	builder := scs.NewBuilder[constraint.U32]
// 	ccs, err := frontend.CompileGeneric[constraint.U32](field.Modulus(), builder, &circuit)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// solve the circuit
// 	twitness, err := frontend.NewWitness(&witness, field.Modulus())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	err = ccs.IsSolved(twitness)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// }

// ------------------------------------------------------------
// test AssertIsCodeWord

type AssertIsCodeWordCircuit struct {
	rate uint64
	d    *fft.Domain
	P    []frontend.Variable `gnark:",public"`
}

func (circuit *AssertIsCodeWordCircuit) Define(api frontend.API) error {
	return assertIsCodeWord(api, circuit.P, circuit.d.GeneratorInv, circuit.d.Cardinality, circuit.rate)

}

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
// EvaluateLagrange

type EvaluateLagrangeCircuit struct {
	P []frontend.Variable
	X frontend.Variable `gnark:",public"`
	R frontend.Variable
	d *fft.Domain
}

func (circuit *EvaluateLagrangeCircuit) Define(api frontend.API) error {

	res := gnarkEvaluateLagrange(api, circuit.P, circuit.X, circuit.d.Generator, circuit.d.Cardinality)

	api.AssertIsEqual(res, circuit.R)

	return nil
}

// func TestEvaluateLagrangeCircuit(t *testing.T) {

// 	// generate witness
// 	size := 16
// 	p := make([]field.Element, size)
// 	for i := 0; i < size; i++ {
// 		p[i].SetRandom()
// 	}
// 	var x fext.Element
// 	x.SetRandom()
// 	var r fext.Element
// 	for i := 0; i < size; i++ {
// 		r.Mul(&r, &x)
// 		var temp fext.Element
// 		fext.FromBase(&temp, &p[len(p)-1-i])
// 		r.Add(&r, &temp)
// 	}

// 	d := fft.NewDomain(uint64(size))
// 	d.FFT(p, fft.DIF)
// 	fft.BitReverse(p)

// 	var witness EvaluateLagrangeCircuit
// 	witness.P = make([]frontend.Variable, size)
// 	for i := 0; i < size; i++ {
// 		witness.P[i] = p[i].String()
// 	}
// 	witness.X = x
// 	witness.R = r

// 	// compile the circuit
// 	var circuit EvaluateLagrangeCircuit
// 	circuit.P = make([]frontend.Variable, size)
// 	circuit.d = d
// 	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// solve the circuit
// 	twitness, err := frontend.NewWitness(&witness, ecc.BLS12_377.ScalarField())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	err = ccs.IsSolved(twitness)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// }

// ------------------------------------------------------------
// N Commitments with Merkle opening

func getProofVortexNCommitmentsWithMerkleNoSis(t *testing.T, nCommitments, nPolys, polySize, blowUpFactor int) (
	proof *OpeningProof,
	randomCoin fext.Element,
	x fext.Element,
	yLists [][]fext.Element,
	entryList []int,
	roots []field.Octuplet,
) {

	x = fext.RandomElement()
	randomCoin = fext.RandomElement()
	entryList = []int{1, 5, 19, 645}

	params := NewParams(blowUpFactor, polySize, nPolys*nCommitments, ringsis.StdParams, poseidon2.NewMerkleDamgardHasher, nil)

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
	roots = make([]field.Octuplet, nCommitments)
	trees := make([]*smt.Tree, nCommitments)
	committedMatrices := make([]EncodedMatrix, nCommitments)
	isSISReplacedByPoseidon2 := make([]bool, nCommitments)
	for j := range trees {
		// As Gnark does not support SIS, we commit without SIS hashing
		committedMatrices[j], trees[j], _ = params.CommitMerkleWithoutSIS(polyLists[j])
		roots[j] = trees[j].Root
		// We set the SIS replaced by Poseidon2 to true, as Gnark does not support SIS
		isSISReplacedByPoseidon2[j] = true
	}

	// Generate the proof
	proof = params.InitOpeningWithLC(utils.Join(polyLists...), randomCoin)
	proof.Complete(entryList, committedMatrices, trees)

	// Check the proof
	err := VerifyOpening(&VerifierInputs{
		Params:                   *params,
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

// func TestGnarkVortexNCommitmentsWithMerkleNoSis(t *testing.T) {

// 	// generate witness
// 	nCommitments := 4
// 	nPolys := 15
// 	polySize := 1 << 10
// 	blowUpFactor := 2
// 	proof, randomCoin, x, ys, entryList, commitments := getProofVortexNCommitmentsWithMerkleNoSis(t, nCommitments, nPolys, polySize, blowUpFactor)

// 	rsSize := blowUpFactor * polySize
// 	rsDomainSize := uint64(rsSize)
// 	rsDomain := fft.NewDomain(rsDomainSize)
// 	var witness VerifyOpeningCircuitMerkleTree
// 	witness.Proof.RsDomain = rsDomain
// 	witness.Proof.Rate = uint64(blowUpFactor)
// 	AllocateCircuitVariablesWithMerkleTree(&witness, *proof, ys, entryList, commitments)
// 	AssignCicuitVariablesWithMerkleTree(&witness, *proof, ys, entryList, commitments)
// 	witness.RandomCoin = randomCoin.String()
// 	witness.X = x.String()
// 	witness.Params.HasherFunc = makeMimcHasherfunc
// 	witness.Params.NoSisHasher = makeMimcHasherfunc

// 	// compile the circuit
// 	var circuit VerifyOpeningCircuitMerkleTree
// 	circuit.Proof.LinearCombination = make([]frontend.Variable, rsSize)
// 	circuit.Proof.Rate = uint64(blowUpFactor)
// 	circuit.Proof.RsDomain = rsDomain
// 	circuit.Params.HasherFunc = makeMimcHasherfunc
// 	circuit.Params.NoSisHasher = makeMimcHasherfunc

// 	AllocateCircuitVariablesWithMerkleTree(&circuit, *proof, ys, entryList, commitments)
// 	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, &circuit, frontend.IgnoreUnconstrainedInputs())
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// solve the circuit
// 	twitness, err := frontend.NewWitness(&witness, ecc.BLS12_377.ScalarField())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	err = ccs.IsSolved(twitness)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

func makePoseidonHasher(api frontend.API) (hash.FieldHasher, error) {
	panic("not implemented")
	// h, err := gposeidon2.NewMerkleDamgardHasher(api)
	// return &h, err
}
