package gnarkutil

import (
	"slices"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
	"github.com/consensys/linea-monorepo/prover/utils"
)

/*
Allocate a slice of field element
*/
// func AllocateSlice(n int) []zk.WrappedVariable {
// 	return make([]zk.WrappedVariable, n)
// }

/*
AllocateSliceExt allocates a slice of field extension elements
*/
func AllocateSliceExt(n int) []gnarkfext.E4Gen {
	return make([]gnarkfext.E4Gen, n)
}

// AsWitnessBls converts a slice of field elements to a slice of witness variables
// of the same length with only public inputs.
func AsWitnessBls(v []zk.WrappedVariable) witness.Witness {

	var (
		wit, _  = witness.New(ecc.BLS12_377.ScalarField())
		witChan = make(chan any, len(v))
	)

	for _, w := range v {
		witChan <- w.AsEmulated()
	}

	close(witChan)

	if err := wit.Fill(len(v), 0, witChan); err != nil {
		panic(err)
	}

	return wit
}

// AsWitnessKoala converts a slice of base field elements to a slice
// of witness variables of the same length with only public inputs. The function
// assumes the WrappedVariable are not emulated element.
func AsWitnessKoala(v []zk.WrappedVariable) witness.Witness {

	var (
		wit, _  = witness.New(koalabear.Modulus())
		witChan = make(chan any)
	)

	go func() {
		for _, w := range v {
			witChan <- w.AsNative()
		}
		close(witChan)
	}()

	if err := wit.Fill(len(v), 0, witChan); err != nil {
		panic(err)
	}

	return wit
}

// EmulatedFromLimbSlice converts slice of frontend.Variable representing inputs
// lower than 2**bitWidth to slice of emulated.Element for the target field. The
// input is expected to be in BIG-ENDIAN form.
func EmulatedFromLimbSlice[T emulated.FieldParams](
	api frontend.API,
	f *emulated.Field[T],
	input []frontend.Variable,
	bitWidth int,
) *emulated.Element[T] {

	input = slices.Clone(input)
	slices.Reverse(input)

	targetNbLimbs, targetBitWidth := emulated.GetEffectiveFieldParams[T](
		api.Compiler().Field(),
	)

	if targetNbLimbs*targetBitWidth < uint(len(input)*bitWidth) {
		utils.Panic(
			"can't fit on emulated field expected#bits=%v provided#bits=%v",
			targetNbLimbs*targetBitWidth, len(input)*bitWidth,
		)
	}

	// Then, it's nice and we don't need to reslice the inputs. We might need to
	// zero pad on the right
	if targetBitWidth == uint(bitWidth) {

		res := make([]frontend.Variable, targetNbLimbs)
		for i := 0; i < len(input); i++ {
			res[i] = input[i]
		}

		for i := len(input); i < int(targetNbLimbs); i++ {
			res[i] = 0
		}

		return f.NewElement(res)
	}

	// Otherwise, we need to slice the inputs. For now, we do it with binary
	// decomposition but it could be optimized. If too few bits are provided,
	// right-pad [bits] with zeroes.
	bits := make([]frontend.Variable, targetNbLimbs*targetBitWidth)
	for i := 0; i < len(input); i++ {
		inputBits := api.ToBinary(input[i], bitWidth)
		for j := range inputBits {
			bits[i*bitWidth+j] = inputBits[j]
		}
	}

	for i := len(input) * bitWidth; i < len(bits); i++ {
		bits[i] = 0
	}

	recomposed := make([]frontend.Variable, targetNbLimbs)
	for i := 0; i < len(recomposed); i++ {
		recomposed[i] = api.FromBinary(bits[i*int(targetBitWidth) : (i+1)*int(targetBitWidth)]...)
	}

	return f.NewElement(recomposed)
}
