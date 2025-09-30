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
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
)

var (
	ErrPNotOfSizeCardinality = errors.New("p should be of size cardinality")
)

// Register fft inverse hint
func init() {
	solver.RegisterHint(fftInverseKoalaBearNative)
}

func FFTInverseKoalaBear[T zk.Element]() solver.Hint {
	var t T
	switch any(t).(type) {
	case zk.EmulatedElement:
		return fftInverseKoalaBearEmulated
	case zk.NativeElement:
		return fftInverseKoalaBearNative
	default:
		panic("unsupported requested API type")
	}
}

func fftInverseKoalaBearEmulated(_ *big.Int, inputs []*big.Int, output []*big.Int) error {
	return emulated.UnwrapHint(inputs, output, fftInverseKoalaBearNative)
}

// fftInverseKoalaBearNative hint for the inverse FFT on koalabear
func fftInverseKoalaBearNative(_ *big.Int, inputs []*big.Int, results []*big.Int) error {

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

// computes fft^-1(p) where the fft is done on <generator>, a set of size cardinality.
// It is assumed that p is correctly sized.
func FFTInverse[T zk.Element](api frontend.API, p []*T, genInv field.Element, cardinality uint64) ([]*T, error) {

	var cardInverse field.Element
	cardInverse.SetUint64(cardinality).Inverse(&cardInverse)

	apiGen, err := zk.NewApi[T](api)
	if err != nil {
		return nil, err
	}

	// res of the fft inverse
	res, err := apiGen.NewHint(FFTInverseKoalaBear[T](), len(p), p...)
	if err != nil {
		return nil, err
	}

	// probabilistically check the result of the FFT
	// multicommit.WithCommitment(
	// 	api,
	// 	func(api frontend.API, x T) error {
	// 		// evaluation canonical
	// 		ec := gnarkEvalCanonical(api, res, x)

	// 		// evaluation Lagrange
	// 		var gen field.Element
	// 		gen.Inverse(&genInv)
	// 		lagranges := gnarkComputeLagrangeAtZ(api, x, gen, cardinality)
	// 		var el T
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
func gnarkEvalCanonical[T zk.Element](api frontend.API, p []*T, z *T) *T {

	apiGen, err := zk.NewApi[T](api)
	if err != nil {
		panic(err)
	}

	res := zk.ValueOf[T](0)
	s := len(p)
	for i := 0; i < len(p); i++ {
		res = apiGen.Mul(res, z)
		res = apiGen.Add(res, p[s-1-i])
	}
	return res
}

func gnarkEvaluateLagrange[T zk.Element](api frontend.API, p []*T, z T, gen field.Element, cardinality uint64) *T {

	apiGen, err := zk.NewApi[T](api)
	if err != nil {
		panic(err)
	}

	res := zk.ValueOf[T](0)

	lagranges := gnarkComputeLagrangeAtZ(api, &z, gen, cardinality)
	for i := uint64(0); i < cardinality; i++ {
		tmp := apiGen.Mul(lagranges[i], p[i])
		res = apiGen.Add(res, tmp)
	}

	return res
}

// computeLagrange returns Lᵢ(ζ) for i=1..n
// with lᵢ(ζ) = ωⁱ/n*(ζⁿ-1)/(ζ - ωⁱ)
// (the g stands for gnark)
func gnarkComputeLagrangeAtZ[T zk.Element](api frontend.API, z *T, gen field.Element, cardinality uint64) []*T {

	apiGen, err := zk.NewApi[T](api)
	if err != nil {
		panic(err)
	}

	res := make([]*T, cardinality)
	tb := bits.TrailingZeros(uint(cardinality))

	// ζⁿ-1
	res[0] = z
	for i := 0; i < tb; i++ {
		res[0] = apiGen.Mul(res[0], res[0])
	}
	res[0] = apiGen.Sub(res[0], zk.ValueOf[T](1))

	// ζ-1
	accZetaMinusOmegai := apiGen.Sub(z, zk.ValueOf[T](1))

	// (ζⁿ-1)/(ζ-1)
	res[0] = apiGen.Div(res[0], accZetaMinusOmegai)

	// 1/n*(ζⁿ-1)/(ζ-1)
	res[0] = apiGen.Div(res[0], zk.ValueOf[T](cardinality))

	// res[i] <- res[i-1] * (ζ-ωⁱ⁻¹)/(ζ-ωⁱ) * ω
	var accOmega field.Element
	accOmega.SetOne()

	for i := uint64(1); i < cardinality; i++ {
		res[i] = apiGen.Mul(res[i-1], zk.ValueOf[T](gen))           // res[i] <- ω * res[i-1]
		res[i] = apiGen.Mul(res[i], accZetaMinusOmegai)             // res[i] <- res[i]*(ζ-ωⁱ⁻¹)
		accOmega.Mul(&accOmega, &gen)                               // accOmega <- accOmega * ω
		accZetaMinusOmegai = apiGen.Sub(z, zk.ValueOf[T](accOmega)) // accZetaMinusOmegai <- ζ-ωⁱ
		res[i] = apiGen.Div(res[i], accZetaMinusOmegai)             // res[i]  <- res[i]/(ζ-ωⁱ)
	}

	return res
}

// Checks that p is a polynomial of degree < cardinality/rate
// * p polynomial of size cardinality
// * genInv inverse of the generator of the subgroup of size cardinality
// * rate of the RS code
func assertIsCodeWord[T zk.Element](api frontend.API, p []*T, genInv koalabear.Element, cardinality, rate uint64) error {

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
		api.AssertIsEqual(pCanonical[i], 0)
	}

	return nil
}

// Opening proof without Merkle proofs
type GProofWoMerkle[T zk.Element] struct {

	// columns on against which the linear combination is checked
	// (the i-th entry is the EntryList[i]-th column). The columns may
	// as well be dispatched in several matrices.
	// Columns [i][j][k] returns the k-th entry of the j-th selected
	// column of the i-th commitment
	Columns [][][]T

	// domain of the RS code
	RsDomain *fft.Domain

	// Rate of the RS code, Blowup factor in Vortex, inverse rate to be precise
	Rate uint64

	// Linear combination of the rows of the polynomial P written as a square matrix
	LinearCombination []*T
}

// Opening proof with Merkle proofs
type GProof[T zk.Element] struct {
	GProofWoMerkle[T]
	MerkleProofs [][]smt.GnarkProof
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

func GnarkVerifyCommon[T zk.Element](
	api frontend.API,
	params GParams,
	proof GProofWoMerkle[T],
	x T,
	ys [][]*T,
	randomCoin T,
	entryList []T,
) ([][]T, error) {

	// check the linear combination is a codeword
	api.Compiler().Defer(func(api frontend.API) error {
		return assertIsCodeWord[T](
			api,
			proof.LinearCombination,
			proof.RsDomain.GeneratorInv,
			proof.RsDomain.Cardinality,
			proof.Rate,
		)
	})

	// Check the consistency of Ys and proof.Linearcombination
	yjoined := utils.Join(ys...)
	alphaY := gnarkEvaluateLagrange[T](
		api,
		proof.LinearCombination,
		x,
		proof.RsDomain.Generator,
		proof.RsDomain.Cardinality)
	alphaYPrime := gnarkEvalCanonical[T](api, yjoined, &randomCoin)
	api.AssertIsEqual(alphaY, alphaYPrime)

	// Size of the hash of 1 column
	numRounds := len(ys)

	selectedColSisDigests := make([][]T, numRounds)
	tbl := logderivlookup.New(api)
	for i := range proof.LinearCombination {
		tbl.Insert(proof.LinearCombination[i])
	}
	for j, selectedColID := range entryList {

		// Will carry the concatenation of the columns for the same entry j
		fullCol := []*T{}

		for i := range selectedColSisDigests {

			if j == 0 {
				selectedColSisDigests[i] = make([]T, len(entryList))
			}

			// Entries of the selected columns #j contained in the commitment #i.
			selectedSubCol := proof.Columns[i][j]
			selectedSubColPtr := make([]*T, len(selectedSubCol))
			for k := 0; k < len(selectedSubCol); k++ {
				selectedSubColPtr[k] = &selectedSubCol[k]
			}
			fullCol = append(fullCol, selectedSubColPtr...)

			// Check consistency between the opened column and the commitment
			if !params.HasNoSisHasher() {
				panic("the vortex verifier circuit only supports a no-SIS hasher")
			}

			// TODO @thomas fixme
			// hasher, _ := params.NoSisHasher(api)
			// hasher.Reset()
			// hasher.Write(selectedSubCol...)
			// digest := hasher.Sum()
			// selectedColSisDigests[i][j] = digest
		}

		// Check the linear combination is consistent with the opened column
		y := gnarkEvalCanonical(api, fullCol, &randomCoin)
		v := tbl.Lookup(selectedColID)[0]
		api.AssertIsEqual(y, v)

	}
	return selectedColSisDigests, nil
}
