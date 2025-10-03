package modexp

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/evmprecompiles"
	"github.com/consensys/gnark/std/math/bitslice"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/gnark/std/math/emulated/emparams"
	"github.com/consensys/linea-monorepo/prover/utils"
)

const (
	// bit-size bound for the operands in the small
	smallModexpSize = 256
	largeModexpSize = 4096
	// limbSize is the size (in bits) of a limb as in the public inputs of the
	// circuit. This is a parameter linked to how the arithmetization encodes
	// 256 bits integers.
	limbSizeBits = 128
)

// ModExpCircuit implements the [frontend.Circuit] interface and is responsible
// for ensuring all the modexp claims brought to the antichamber module.
//
// The circuit is meant to be used in two variants:
//   - 256 bits, where all the operands and the claimed result have a size
//     smaller than 256 bits.
//   - 4096, where the operands are bound to 4096 bits
type ModExpCircuit struct {
	Instances []modexpCircuitInstance `gnark:",public"`
}

// modexpCircuitInstance is a substructre interface and is
// responsible for ensuring the correctness of the evaluation of the MODEXP
// precompile for a single instance of MODEXP.
//
// The operands are represented in limbs of 16 bytes.
type modexpCircuitInstance struct {
	Base     []T `gnark:",public"`
	Exponent []T `gnark:",public"`
	Modulus  []T `gnark:",public"`
	Result   []T `gnark:",public"`
}

// allocate256Bits allocates [ModExpCircuit] for n instances assuming the 256-bit
// variant.
func allocateCircuit(n int, numBits int) *ModExpCircuit {

	if numBits != smallModexpSize && numBits != largeModexpSize {
		utils.Panic("expected `numBits = {%v, %v}`", smallModexpSize, largeModexpSize)
	}

	var (
		numLimbs = numBits / limbSizeBits
		res      = &ModExpCircuit{
			Instances: make([]modexpCircuitInstance, n),
		}
	)

	for i := range res.Instances {
		res.Instances[i].Base = make([]T, numLimbs)
		res.Instances[i].Exponent = make([]T, numLimbs)
		res.Instances[i].Modulus = make([]T, numLimbs)
		res.Instances[i].Result = make([]T, numLimbs)
	}

	return res
}

// Define implements the [frontend.Circuit] interface
func (m *ModExpCircuit) Define(api frontend.API) error {

	for i := range m.Instances {

		var (
			instance = m.Instances[i]
			numLimbs = len(m.Instances[i].Base)
		)

		switch numLimbs * limbSizeBits {
		case smallModexpSize:
			checkModexpInstance[emparams.Mod1e256](api, &instance)
		case largeModexpSize:
			checkModexpInstance[emparams.Mod1e4096](api, &instance)
		default:
			utils.Panic(
				"Unexpected field size = %v, should be either %v or %v",
				numLimbs*limbSizeBits, smallModexpSize, largeModexpSize,
			)
		}
	}

	return nil
}

// defined4096 implements the circuit logic for the case where
func checkModexpInstance[P emulated.FieldParams](api frontend.API, m *modexpCircuitInstance) {

	var (
		params        P
		emApi, errAPI = emulated.NewField[P](api)
		baseLimbs     = make([]T, params.NbLimbs())
		exponentLimbs = make([]T, params.NbLimbs())
		modulusLimbs  = make([]T, params.NbLimbs())
		resultLimbs   = make([]T, params.NbLimbs())
	)

	if errAPI != nil {
		utils.Panic("could not generate the API: %v", errAPI)
	}

	for i := range m.Base {
		// The limbs are provided by the arithmetization in big-endian order
		// but the binary API of gnark manages bits in little-endian order. So
		// we have to account for this when unpacking all the bits of the
		// operands.
		posInLimbs := len(m.Base) - i - 1
		baseLimbs[2*posInLimbs], baseLimbs[2*posInLimbs+1] = bitslice.Partition(api, m.Base[i], params.BitsPerLimb(), bitslice.WithNbDigits(limbSizeBits))
		exponentLimbs[2*posInLimbs], exponentLimbs[2*posInLimbs+1] = bitslice.Partition(api, m.Exponent[i], params.BitsPerLimb(), bitslice.WithNbDigits(limbSizeBits))
		modulusLimbs[2*posInLimbs], modulusLimbs[2*posInLimbs+1] = bitslice.Partition(api, m.Modulus[i], params.BitsPerLimb(), bitslice.WithNbDigits(limbSizeBits))
		resultLimbs[2*posInLimbs], resultLimbs[2*posInLimbs+1] = bitslice.Partition(api, m.Result[i], params.BitsPerLimb(), bitslice.WithNbDigits(limbSizeBits))
	}

	var (
		base           = emApi.NewElement(baseLimbs)
		exponent       = emApi.NewElement(exponentLimbs)
		modulus        = emApi.NewElement(modulusLimbs)
		resultExpected = emApi.NewElement(resultLimbs)
		resultActual   = evmprecompiles.Expmod(api, base, exponent, modulus)
	)

	emApi.AssertIsEqual(resultExpected, resultActual)
}
