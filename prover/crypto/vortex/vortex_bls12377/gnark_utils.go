package vortex

import (
	"errors"
	"fmt"
	"math/big"
	"math/bits"

	gnarkbits "github.com/consensys/gnark/std/math/bits"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/cmp"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_bls12377"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_bls12377"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
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
	//TODO @thomas: how to implement the emulated version
	return emulated.UnwrapHint(inputs, output, fftInverseNative)
}

// fftInverse hint for the inverse FFT on koalabear
func fftInverseNative(_ *big.Int, inputs []*big.Int, results []*big.Int) error {

	if len(inputs)%4 != 0 || len(results) != len(inputs) {
		return fmt.Errorf("inputs length must be a multiple of 4 and results length must match inputs length")
	}
	// TODO store this somewhere (global variable or something, shouldn't regenerate it at each call)
	d := fft.NewDomain(uint64(len(inputs)/4), fft.WithoutPrecompute())

	v := make([]fext.Element, len(inputs)/4)
	for i := 0; i < len(inputs)/4; i++ {
		v[i].B0.A0.SetBigInt(inputs[4*i])
		v[i].B0.A1.SetBigInt(inputs[4*i+1])
		v[i].B1.A0.SetBigInt(inputs[4*i+2])
		v[i].B1.A1.SetBigInt(inputs[4*i+3])
	}
	d.FFTInverseExt(v, fft.DIF)
	fft.BitReverse(v)

	for i := 0; i < len(results)/4; i++ {
		results[4*i] = big.NewInt(0)
		results[4*i+1] = big.NewInt(0)
		results[4*i+2] = big.NewInt(0)
		results[4*i+3] = big.NewInt(0)

		v[i].B0.A0.BigInt(results[4*i])
		v[i].B0.A1.BigInt(results[4*i+1])
		v[i].B1.A0.BigInt(results[4*i+2])
		v[i].B1.A1.BigInt(results[4*i+3])
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
func FFTInverseExt(api frontend.API, p []gnarkfext.E4Gen, genInv field.Element, cardinality uint64) ([]gnarkfext.E4Gen, error) {

	var cardInverse field.Element
	cardInverse.SetUint64(cardinality).Inverse(&cardInverse)

	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		return []gnarkfext.E4Gen{}, err
	}

	res, err := ext4.NewHint(fftInverseHint(ext4.ApiGen.Type()), len(p), p...)

	if err != nil {
		return nil, err
	}

	// probabilistically check the result of the FFT
	// TODO @thomas TODO@yao: how to fix this test with 	multicommit.WithCommitment

	// Generate randomness for probabilistic verification
	xGen4, err := computeRandomness(api, p)

	if err != nil {
		return []gnarkfext.E4Gen{}, err
	}

	ec := gnarkEvalCanonicalExt(api, res, xGen4)

	// evaluation Lagrange
	var gen field.Element
	gen.Inverse(&genInv)
	lagranges := gnarkComputeLagrangeAtZ(api, xGen4, gen, cardinality)
	el := ext4.Zero()
	for i := 0; i < len(p); i++ {
		tmp := ext4.Mul(&p[i], &lagranges[i])
		el = ext4.Add(el, tmp)
	}

	ext4.AssertIsEqual(&ec, el)

	return res, nil

}

// Computes randomness used for probabilistic verification.
// Generates xGen4 by hashing `p` using Poseidon hashers (different implementations for native vs emulated fields).
func computeRandomness(api frontend.API, p []gnarkfext.E4Gen) (gnarkfext.E4Gen, error) {

	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		return gnarkfext.E4Gen{}, err
	}

	xGen4 := gnarkfext.E4Gen{}

	if ext4.ApiGen.Type() == zk.Native {
		h, err := poseidon2_koalabear.NewGnarkMDHasher(api)
		if err != nil {
			return gnarkfext.E4Gen{}, err
		}

		for i := range p {
			h.Write(p[i].B0.A0.AsNative())
			h.Write(p[i].B0.A1.AsNative())
			h.Write(p[i].B1.A0.AsNative())
			h.Write(p[i].B1.A1.AsNative())
		}
		xOct := h.Sum()

		xGen4.B0.A0 = zk.WrapFrontendVariable(xOct[0])
		xGen4.B0.A1 = zk.WrapFrontendVariable(xOct[1])
		xGen4.B1.A0 = zk.WrapFrontendVariable(xOct[2])
		xGen4.B1.A1 = zk.WrapFrontendVariable(xOct[3])
	} else {
		h, err := poseidon2_bls12377.NewGnarkMDHasher(api)
		if err != nil {
			return gnarkfext.E4Gen{}, err
		}

		for i := range p {
			h.Write(p[i].B0.A0.AsNative())
			h.Write(p[i].B0.A1.AsNative())
			h.Write(p[i].B1.A0.AsNative())
			h.Write(p[i].B1.A1.AsNative())
		}
		xOct := EncodeFVTo8WVs(api, h.Sum())

		xGen4.B0.A0 = xOct[0]
		xGen4.B0.A1 = xOct[1]
		xGen4.B1.A0 = xOct[2]
		xGen4.B1.A1 = xOct[3]
	}
	return xGen4, nil
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

func gnarkEvaluateLagrangeExt(api frontend.API, p []gnarkfext.E4Gen, z gnarkfext.E4Gen, gen field.Element, cardinality uint64) gnarkfext.E4Gen {

	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
	}
	res := *ext4.Zero()
	lagranges := gnarkComputeLagrangeAtZ(api, z, gen, cardinality)

	for i := uint64(0); i < cardinality; i++ {
		tmp := ext4.Mul(&lagranges[i], &p[i])
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

func assertIsCodeWordExt(api frontend.API, p []gnarkfext.E4Gen, genInv koalabear.Element, cardinality, rate uint64) error {
	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
	}

	if uint64(len(p)) != cardinality {
		return ErrPNotOfSizeCardinality
	}

	// get the canonical form of p
	pCanonical, err := FFTInverseExt(api, p, genInv, cardinality)
	if err != nil {
		return err
	}

	// check that is of degree < cardinality/rate
	degree := (cardinality - (cardinality % rate)) / rate
	for i := degree; i < cardinality; i++ {
		zeroValue := gnarkfext.NewE4Gen(fext.Zero())
		ext4.AssertIsEqual(&pCanonical[i], &zeroValue)
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
	LinearCombination []gnarkfext.E4Gen
}

// Opening proof with Merkle proofs
type GProof struct {
	GProofWoMerkle
	MerkleProofs [][]smt_bls12377.GnarkProof
}

// Gnark params
type GParams struct {
	Key         *ringsis.Key
	HasherFunc  func(frontend.API) (poseidon2_bls12377.GnarkMDHasher, error)
	NoSisHasher func(frontend.API) (poseidon2_bls12377.GnarkMDHasher, error)
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
	entryList []zk.WrappedVariable,
) ([][]frontend.Variable, error) {

	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		panic(err)
	}

	assertIsCodeWordExt(
		api,
		proof.LinearCombination,
		proof.RsDomain.GeneratorInv,
		proof.RsDomain.Cardinality,
		proof.Rate,
	)

	// Check the consistency of Ys and proof.Linearcombination
	yjoined := utils.Join(ys...)
	alphaY := gnarkEvaluateLagrangeExt(
		api,
		proof.LinearCombination,
		x,
		proof.RsDomain.Generator,
		proof.RsDomain.Cardinality)
	alphaYPrime := gnarkEvalCanonicalExt(api, yjoined, randomCoin)
	ext4.AssertIsEqual(&alphaY, &alphaYPrime)

	// Size of the hash of 1 column
	numRounds := len(ys)
	selectedColSisDigests := make([][]frontend.Variable, numRounds)

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
			hashinput := EncodeWVsToFVs(api, selectedSubCol)
			hasher.Write(hashinput...)
			digest := hasher.Sum()

			selectedColSisDigests[i][j] = digest
		}

		// Check the linear combination is consistent with the opened column
		y := gnarkEvalCanonical(api, fullCol, randomCoin)
		v := Mux(api, selectedColID.V, proof.LinearCombination...)
		ext4.AssertIsEqual(&y, &v)

	}
	return selectedColSisDigests, nil
}

// Mux is an n to 1 multiplexer: out = inputs[sel]. In other words, it selects
// exactly one of its inputs based on sel. The index of inputs starts from zero.
//
// sel needs to be between 0 and n - 1 (inclusive), where n is the number of
// inputs, otherwise the proof will fail.
func Mux(api frontend.API, sel frontend.Variable, inputs ...gnarkfext.E4Gen) gnarkfext.E4Gen {
	n := uint(len(inputs))
	if n == 0 {
		panic("invalid input length 0 for mux")
	}
	if n == 1 {
		api.AssertIsEqual(sel, 0)
		return inputs[0]
	}
	nbBits := bits.Len(n - 1)                                               // we use n-1 as sel is 0-indexed
	selBits := gnarkbits.ToBinary(api, sel, gnarkbits.WithNbDigits(nbBits)) // binary decomposition ensures sel < 2^nbBits

	// We use BinaryMux when len(inputs) is a power of 2.
	if bits.OnesCount(n) == 1 {
		return BinaryMux(api, selBits, inputs)
	}

	bcmp := cmp.NewBoundedComparator(api, big.NewInt((1<<nbBits)-1), false)
	bcmp.AssertIsLessEq(sel, n-1)

	// Otherwise, we split inputs into two sub-arrays, such that the first part's length is 2's power
	return muxRecursive(api, selBits, inputs)
}

func muxRecursive(api frontend.API,
	selBits []frontend.Variable, inputs []gnarkfext.E4Gen) gnarkfext.E4Gen {

	nbBits := len(selBits)
	leftCount := uint(1 << (nbBits - 1))
	left := BinaryMux(api, selBits[:nbBits-1], inputs[:leftCount])

	rightCount := uint(len(inputs)) - leftCount
	nbRightBits := bits.Len(rightCount)

	var right gnarkfext.E4Gen
	if bits.OnesCount(rightCount) == 1 {
		right = BinaryMux(api, selBits[:nbRightBits-1], inputs[leftCount:])
	} else {
		right = muxRecursive(api, selBits[:nbRightBits], inputs[leftCount:])
	}

	msb := selBits[nbBits-1]

	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		return gnarkfext.E4Gen{}
	}

	return *ext4.Select(zk.WrapFrontendVariable(msb), &right, &left)
}
func BinaryMux(api frontend.API, selBits []frontend.Variable, inputs []gnarkfext.E4Gen) gnarkfext.E4Gen {
	if len(inputs) != 1<<len(selBits) {
		panic(fmt.Sprintf("invalid input length for BinaryMux (%d != 2^%d)", len(inputs), len(selBits)))
	}

	for _, b := range selBits {
		api.AssertIsBoolean(b)
	}

	return binaryMuxRecursive(api, selBits, inputs)
}

func binaryMuxRecursive(api frontend.API, selBits []frontend.Variable, inputs []gnarkfext.E4Gen) gnarkfext.E4Gen {
	ext4, err := gnarkfext.NewExt4(api)
	if err != nil {
		return gnarkfext.E4Gen{}
	}
	// The number of defined R1CS constraints for an input of length n is always n - 1.
	// n does not need to be a power of 2.
	if len(selBits) == 0 {
		return inputs[0]
	}

	nextSelBits := selBits[:len(selBits)-1]
	msb := selBits[len(selBits)-1]
	pivot := 1 << len(nextSelBits)
	if pivot >= len(inputs) {
		return binaryMuxRecursive(api, nextSelBits, inputs)
	}

	left := binaryMuxRecursive(api, nextSelBits, inputs[:pivot])
	right := binaryMuxRecursive(api, nextSelBits, inputs[pivot:])
	return *ext4.Add(&left, ext4.MulByFp(ext4.Sub(&right, &left), zk.WrapFrontendVariable(msb)))
}
