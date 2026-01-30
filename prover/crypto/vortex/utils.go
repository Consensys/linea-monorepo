package vortex

import (
	"errors"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/utils"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/koalagnark"
)

func init() {
	solver.RegisterHint(
		fftExtInvHintEmulated, fftExtInvHintNative,
	)
}

var (
	ErrSizeNotAMultipleOfFour = errors.New("size(inputs) should be a multiple of 4")
	ErrSizesDontMatch         = errors.New("size(inputs) should be equal to size(outputs)")
	ErrNotAPowerOfTwo         = errors.New("size(inputs) should be a power of two")
)

func fftInvHint(t koalagnark.VType) solver.Hint {
	if t == koalagnark.Native {
		return fftExtInvHintNative
	} else {
		return fftExtInvHintEmulated
	}
}

func fftExtInvHintEmulated(_ *big.Int, inputs []*big.Int, output []*big.Int) error {
	return emulated.UnwrapHint(inputs, output, fftExtInvHintNative)
}

// Each chunk of 4 inputs corresponds to a E4 element.
func fftExtInvHintNative(scalarField *big.Int, inputs, outputs []*big.Int) error {

	if len(inputs)%4 != 0 {
		return ErrSizeNotAMultipleOfFour
	}

	if len(inputs) != len(outputs) {
		return ErrSizesDontMatch
	}
	n := len(inputs) / 4

	_n := ecc.NextPowerOfTwo(uint64(n))
	if _n != uint64(n) {
		return ErrNotAPowerOfTwo
	}

	d := fft.NewDomain(uint64(n))

	_res := make([]fext.Element, n)
	for i := 0; i < n; i++ {
		_res[i].B0.A0.SetBigInt(inputs[4*i])
		_res[i].B0.A1.SetBigInt(inputs[4*i+1])
		_res[i].B1.A0.SetBigInt(inputs[4*i+2])
		_res[i].B1.A1.SetBigInt(inputs[4*i+3])
	}
	d.FFTInverseExt(_res, fft.DIF)
	utils.BitReverse(_res)

	for i := 0; i < n; i++ {
		_res[i].B0.A0.BigInt(outputs[4*i])
		_res[i].B0.A1.BigInt(outputs[4*i+1])
		_res[i].B1.A0.BigInt(outputs[4*i+2])
		_res[i].B1.A1.BigInt(outputs[4*i+3])
	}

	return nil
}
