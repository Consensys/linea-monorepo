package vortex

import (
	"errors"
	"math/big"
	"math/bits"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash"
	"github.com/consensys/gnark/std/lookup/logderivlookup"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_bls12377"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
)

var (
	ErrPNotOfSizeCardinality = errors.New("p should be of size cardinality")
)

// Register fft inverse hint
func init() {
	solver.RegisterHint(fftInverseNative, fftInverseEmulated)
}

func fftInverseHint(t zk.VType) solver.Hint {
	if t == zk.Native {
		return fftInverseNative
	} else {
		return fftInverseEmulated
	}
}

func fftInverseEmulated(_ *big.Int, inputs []*big.Int, output []*big.Int) error {
	return emulated.UnwrapHint(inputs, output, fftInverseNative)
}

// fftInverse hint for the inverse FFT on koalabear
func fftInverseNative(_ *big.Int, inputs []*big.Int, results []*big.Int) error {

	// TODO store this somewhere (global variable or something, shouldn't regenerate it at each call)
	d := fft.NewDomain(uint64(len(inputs)), fft.WithoutPrecompute())

	v := make([]field.Element, len(inputs))
	for i := 0; i < len(inputs); i++ {
		v[i].SetBigInt(inputs[i])
	}
	d.FFTInverse(v, fft.DIF)
	fft.BitReverse(v)

	for i := 0; i < len(results); i++ {
		results[i] = big.NewInt(0)
		v[i].BigInt(results[i])
	}

	return nil
}

func toPtr(src []zk.WrappedVariable) []*zk.WrappedVariable {
	res := make([]*zk.WrappedVariable, len(src))
	for i := 0; i < len(res); i++ {
		res[i] = &src[i]
	}
	return res
}

func fromPtr(src []*zk.WrappedVariable) []zk.WrappedVariable {
	res := make([]zk.WrappedVariable, len(src))
	for i := 0; i < len(res); i++ {
		res[i] = *src[i]
	}
	return res
}

// computes fft^-1(p) where the fft is done on <generator>, a set of size cardinality.
// It is assumed that p is correctly sized.
func FFTInverse(api frontend.API, p []zk.WrappedVariable, genInv field.Element, cardinality uint64) ([]zk.WrappedVariable, error) {

	var cardInverse field.Element
	cardInverse.SetUint64(cardinality).Inverse(&cardInverse)

	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		return []zk.WrappedVariable{}, err
	}

	// res of the fft inverse
	res, err := apiGen.NewHint(fftInverseHint(apiGen.Type()), len(p), p...)
	if err != nil {
		return nil, err
	}
	// probabilistically check the result of the FFT
	// multicommit.WithCommitment(
	// 	api,
	// 	func(api frontend.API, x zk.WrappedVariable) error {
	// 		// evaluation canonical
	// 		ec := gnarkEvalCanonical(api, res, x)

	// 		// evaluation Lagrange
	// 		var gen field.Element
	// 		gen.Inverse(&genInv)
	// 		lagranges := gnarkComputeLagrangeAtZ(api, x, gen, cardinality)
	// 		var el zk.WrappedVariable
	// 		el = 0
	// 		for i := 0; i < len(p); i++ {
	// 			tmp := api.Mul(p[i], lagranges[i])
	// 			el = api.Add(el, tmp)
	// 		}

	// 		api.AssertIsEqual(ec, el)
	// 		return nil
	// 	},
	// 	p...,
	// )

	return res, nil
}

// gnarkEvalCanonical evaluates p at z where p represents the polnyomial ∑ᵢp[i]Xⁱ
func gnarkEvalCanonical(api frontend.API, p []zk.WrappedVariable, z gnarkfext.E4Gen) gnarkfext.E4Gen {

	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
	}

	res := *ext4.Zero()
	s := len(p)
	for i := 0; i < len(p); i++ {
		res = *ext4.Mul(&res, &z)
		res = *ext4.AddByBase(&res, p[s-1-i])
	}
	return res
}

// gnarkEvalCanonical evaluates p at z where p represents the polnyomial ∑ᵢp[i]Xⁱ
func gnarkEvalCanonicalExt(api frontend.API, p []gnarkfext.E4Gen, z gnarkfext.E4Gen) gnarkfext.E4Gen {

	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
	}

	res := *ext4.Zero()
	s := len(p)
	for i := 0; i < len(p); i++ {
		res = *ext4.Mul(&res, &z)
		res = *ext4.Add(&res, &p[s-1-i])
	}
	return res
}
func gnarkEvaluateLagrange(api frontend.API, p []zk.WrappedVariable, z gnarkfext.E4Gen, gen field.Element, cardinality uint64) gnarkfext.E4Gen {

	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
	}
	res := *ext4.Zero()
	lagranges := gnarkComputeLagrangeAtZ(api, z, gen, cardinality)

	for i := uint64(0); i < cardinality; i++ {
		tmp := ext4.MulByFp(&lagranges[i], p[i])
		res = *ext4.Add(&res, tmp)
	}

	return res
}

// computeLagrange returns Lᵢ(ζ) for i=1..n
// with lᵢ(ζ) = ωⁱ/n*(ζⁿ-1)/(ζ - ωⁱ)
// (the g stands for gnark)
func gnarkComputeLagrangeAtZ(api frontend.API, z gnarkfext.E4Gen, gen field.Element, cardinality uint64) []gnarkfext.E4Gen {

	res := make([]gnarkfext.E4Gen, cardinality)
	tb := bits.TrailingZeros(uint(cardinality))

	ext4, err := gnarkfext.NewExt4(api)

	if err != nil {
		panic(err)
	}

	// ζⁿ-1
	res[0] = z
	for i := 0; i < tb; i++ {
		res[0] = *ext4.Mul(&res[0], &res[0])
	}

	wOne := ext4.One()
	res[0] = *ext4.Sub(&res[0], wOne)

	// ζ-1
	accZetaMinusOmegai := *ext4.Sub(&z, wOne)

	// (ζⁿ-1)/(ζ-1)
	res[0] = *ext4.Div(&res[0], &accZetaMinusOmegai)

	// 1/n*(ζⁿ-1)/(ζ-1)
	wCardinality := zk.ValueOf(cardinality)
	res[0] = *ext4.DivByBase(&res[0], wCardinality)

	// res[i] <- res[i-1] * (ζ-ωⁱ⁻¹)/(ζ-ωⁱ) * ω
	var accOmega field.Element
	accOmega.SetOne()
	wGen := zk.ValueOf(gen)
	var wAccOmega zk.WrappedVariable
	for i := uint64(1); i < cardinality; i++ {
		res[i] = *ext4.MulByFp(&res[i-1], wGen)          // res[i] <- ω * res[i-1]
		res[i] = *ext4.Mul(&res[i], &accZetaMinusOmegai) // res[i] <- res[i]*(ζ-ωⁱ⁻¹)
		accOmega.Mul(&accOmega, &gen)                    // accOmega <- accOmega * ω
		wAccOmega = zk.ValueOf(accOmega.String())
		wAccOmegaExt := gnarkfext.FromBase(wAccOmega)
		accZetaMinusOmegai = *ext4.Sub(&z, &wAccOmegaExt) // accZetaMinusOmegai <- ζ-ωⁱ
		res[i] = *ext4.Div(&res[i], &accZetaMinusOmegai)  // res[i]  <- res[i]/(ζ-ωⁱ)
	}

	return res
}

// Checks that p is a polynomial of degree < cardinality/rate
// * p polynomial of size cardinality
// * genInv inverse of the generator of the subgroup of size cardinality
// * rate of the RS code
func assertIsCodeWord(api frontend.API, p []zk.WrappedVariable, genInv koalabear.Element, cardinality, rate uint64) error {
	apiGen, err := zk.NewGenericApi(api)
	if err != nil {
		panic(err)
	}

	if uint64(len(p)) != cardinality {
		return ErrPNotOfSizeCardinality
	}

	// get the canonical form of p
	pCanonical, err := FFTInverse(api, p, genInv, cardinality)
	if err != nil {
		return err
	}

	// check that is of degree < cardinality/rate
	degree := (cardinality - (cardinality % rate)) / rate
	for i := degree; i < cardinality; i++ {
		apiGen.AssertIsEqual(pCanonical[i], zk.ValueOf(0))
	}

	return nil
}

// Opening proof without Merkle proofs
type GProofWoMerkle struct {

	// columns on against which the linear combination is checked
	// (the i-th entry is the EntryList[i]-th column). The columns may
	// as well be dispatched in several matrices.
	// Columns [i][j][k] returns the k-th entry of the j-th selected
	// column of the i-th commitment
	Columns [][][]zk.WrappedVariable

	// domain of the RS code
	RsDomain *fft.Domain

	// Rate of the RS code, Blowup factor in Vortex, inverse rate to be precise
	Rate uint64

	// Linear combination of the rows of the polynomial P written as a square matrix
	LinearCombination []zk.WrappedVariable
}

// Opening proof with Merkle proofs
type GProof struct {
	GProofWoMerkle
	MerkleProofs [][]smt_bls12377.GnarkProof
}

// Gnark params
type GParams struct {
	Key         *ringsis.Key
	HasherFunc  func(frontend.API) (hash.FieldHasher, error)
	NoSisHasher func(frontend.API) (hash.FieldHasher, error)
}

func (p *GParams) HasNoSisHasher() bool {
	return p.NoSisHasher != nil
}

func GnarkVerifyCommon(
	api frontend.API,
	params GParams,
	proof GProofWoMerkle,
	x gnarkfext.E4Gen,
	ys [][]gnarkfext.E4Gen,
	randomCoin gnarkfext.E4Gen,
	entryList []frontend.Variable,
) ([][]frontend.Variable, error) {

	// apiGen, err := zk.NewGenericApi(api)
	// if err != nil {
	// 	panic(err)
	// }

	// check the linear combination is a codeword
	api.Compiler().Defer(func(api frontend.API) error {
		return assertIsCodeWord(
			api,
			proof.LinearCombination,
			proof.RsDomain.GeneratorInv,
			proof.RsDomain.Cardinality,
			proof.Rate,
		)
	})

	// Check the consistency of Ys and proof.Linearcombination
	yjoined := utils.Join(ys...)
	alphaY := gnarkEvaluateLagrange(
		api,
		proof.LinearCombination,
		x,
		proof.RsDomain.Generator,
		proof.RsDomain.Cardinality)
	alphaYPrime := gnarkEvalCanonicalExt(api, yjoined, randomCoin)
	api.AssertIsEqual(alphaY, alphaYPrime)

	// Size of the hash of 1 column
	numRounds := len(ys)

	selectedColSisDigests := make([][]frontend.Variable, numRounds)
	tbl := logderivlookup.New(api)
	for i := range proof.LinearCombination {
		tbl.Insert(proof.LinearCombination[i])
	}
	for j, selectedColID := range entryList {

		// Will carry the concatenation of the columns for the same entry j
		fullCol := []zk.WrappedVariable{}

		for i := range selectedColSisDigests {

			if j == 0 {
				selectedColSisDigests[i] = make([]frontend.Variable, len(entryList))
			}

			// Entries of the selected columns #j contained in the commitment #i.
			selectedSubCol := proof.Columns[i][j]
			fullCol = append(fullCol, selectedSubCol...)

			// Check consistency between the opened column and the commitment
			if !params.HasNoSisHasher() {
				panic("the vortex verifier circuit only supports a no-SIS hasher")
			}

			hasher, _ := params.NoSisHasher(api)
			hasher.Reset()
			// TODO @thomas fixme
			// TODO@yao: how to write 1 zero.element and then 7 field.Element to hasher? ...

			hashinput := EncodeWVsToFVs(api, selectedSubCol)
			hasher.Write(hashinput...)
			digest := hasher.Sum()
			selectedColSisDigests[i][j] = digest
		}

		// Check the linear combination is consistent with the opened column
		y := gnarkEvalCanonical(api, fullCol, randomCoin)
		v := tbl.Lookup(selectedColID)[0]
		api.AssertIsEqual(y, v)

	}
	return selectedColSisDigests, nil
}
