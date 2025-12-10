package ecdsa

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra"
	"github.com/consensys/gnark/std/algebra/emulated/sw_emulated"
	"github.com/consensys/gnark/std/evmprecompiles"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/gnark/std/math/emulated/emparams"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated/plonk"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/common"
)

type EcRecoverInstance struct {
	PKXHi, PKXLo, PKYHi, PKYLo [common.NbLimbU128]frontend.Variable `gnark:",public"`
	HHi, HLo                   [common.NbLimbU128]frontend.Variable `gnark:",public"`
	VHi, VLo                   [common.NbLimbU128]frontend.Variable `gnark:",public"`
	RHi, RLo                   [common.NbLimbU128]frontend.Variable `gnark:",public"`
	SHi, SLo                   [common.NbLimbU128]frontend.Variable `gnark:",public"`
	SUCCESS_BIT                frontend.Variable                    `gnark:",public"`
	ECRECOVERBIT               frontend.Variable                    `gnark:",public"`
}

// convertBE16x16ToLE26x10 converts big-endian 16×16-bit limbs (Hi, Lo arrays)
// to little-endian 26×10-bit limbs for gnark's emulated field over small fields.
// Hi and Lo are each 8 limbs of 16 bits in big-endian order (index 0 = most significant).
//
// Input layout:
//   - lo[7] contains bits [0-15] of full 256-bit value (LSB chunk)
//   - lo[0] contains bits [112-127]
//   - hi[7] contains bits [128-143]
//   - hi[0] contains bits [240-255] (MSB chunk)
//
// Output: 26 limbs of 10 bits in little-endian order (limbs[0] = LSB).
func convertBE16x16ToLE26x10(api frontend.API, hi, lo [common.NbLimbU128]frontend.Variable) []frontend.Variable {
	// Collect all 256 bits in little-endian order (bit 0 first).
	// Process limbs from LSB to MSB: lo[7], lo[6], ..., lo[0], hi[7], ..., hi[0]
	// ToBinary returns bits in little-endian order (bits[0] = LSB of limb),
	// which is exactly what we need.
	allBits := make([]frontend.Variable, 256)
	bitPos := 0

	// Lo part: lo[7] is LSB chunk (bits 0-15), lo[0] is MSB chunk of Lo (bits 112-127)
	for i := common.NbLimbU128 - 1; i >= 0; i-- {
		bits := api.ToBinary(lo[i], 16)
		for j := 0; j < 16; j++ {
			allBits[bitPos] = bits[j]
			bitPos++
		}
	}

	// Hi part: hi[7] is LSB chunk of Hi (bits 128-143), hi[0] is MSB chunk (bits 240-255)
	for i := common.NbLimbU128 - 1; i >= 0; i-- {
		bits := api.ToBinary(hi[i], 16)
		for j := 0; j < 16; j++ {
			allBits[bitPos] = bits[j]
			bitPos++
		}
	}

	// Now allBits is in little-endian order: allBits[0] = bit 0 (LSB), allBits[255] = bit 255 (MSB)
	// Build 26 limbs of 10 bits each, also in little-endian order
	result := make([]frontend.Variable, 26)
	for i := 0; i < 26; i++ {
		limbBits := make([]frontend.Variable, 10)
		for j := 0; j < 10; j++ {
			bitIdx := i*10 + j
			if bitIdx < 256 {
				limbBits[j] = allBits[bitIdx]
			} else {
				// Padding for bits beyond 256 (most significant limbs)
				limbBits[j] = frontend.Variable(0)
			}
		}
		result[i] = api.FromBinary(limbBits...)
	}

	return result
}

type MultiEcRecoverCircuit struct {
	Instances []EcRecoverInstance `gnark:",public"`
}

func newMultiEcRecoverCircuit(nbInstances int) *MultiEcRecoverCircuit {
	return &MultiEcRecoverCircuit{
		Instances: make([]EcRecoverInstance, nbInstances),
	}
}

func (c *MultiEcRecoverCircuit) Define(api frontend.API) error {
	curve, err := algebra.GetCurve[emparams.Secp256k1Fr, sw_emulated.AffinePoint[emparams.Secp256k1Fp]](api)
	if err != nil {
		return fmt.Errorf("get curve: %w", err)
	}
	for i := 0; i < len(c.Instances); i++ {
		PK, msg, v, r, s, strictRange, isFailure, err := c.Instances[i].splitInputs(api)
		if err != nil {
			return fmt.Errorf("split inputs: %w", err)
		}
		recovered := evmprecompiles.ECRecover(api, *msg, v, *r, *s, strictRange, isFailure)
		curve.AssertIsEqual(PK, recovered)
	}
	return nil
}

func (c *EcRecoverInstance) splitInputs(api frontend.API) (PK *sw_emulated.AffinePoint[emparams.Secp256k1Fp], msg *emulated.Element[emparams.Secp256k1Fr], v frontend.Variable, r, s *emulated.Element[emparams.Secp256k1Fr], strictRange, isFailure frontend.Variable, err error) {
	fr, err2 := emulated.NewField[emparams.Secp256k1Fr](api)
	if err2 != nil {
		err = fmt.Errorf("field emulation: %w", err2)
		return
	}
	fp, err2 := emulated.NewField[emparams.Secp256k1Fp](api)
	if err2 != nil {
		err = fmt.Errorf("field emulation: %w", err2)
		return
	}
	// Convert big-endian 16×16-bit limbs to little-endian 26×10-bit limbs
	// for gnark's emulated field over KoalaBear
	msg = fr.NewElement(convertBE16x16ToLE26x10(api, c.HHi, c.HLo))
	PK = &sw_emulated.AffinePoint[emparams.Secp256k1Fp]{
		X: *fp.NewElement(convertBE16x16ToLE26x10(api, c.PKXHi, c.PKXLo)),
		Y: *fp.NewElement(convertBE16x16ToLE26x10(api, c.PKYHi, c.PKYLo)),
	}
	// v is 27 or 28 in EVM, but arithmetization gives on two limbs. Ensure that all high limbs are zero.
	for i := 0; i < common.NbLimbU128; i++ {
		api.AssertIsEqual(c.VHi[i], 0)
	}
	// Also assert all but the last low limb are zero (v fits in ~5 bits)
	for i := 0; i < common.NbLimbU128-1; i++ {
		api.AssertIsEqual(c.VLo[i], 0)
	}
	v = c.VLo[common.NbLimbU128-1] // last limb since v is small (27 or 28).

	// Convert r and s similarly
	r = fr.NewElement(convertBE16x16ToLE26x10(api, c.RHi, c.RLo))
	s = fr.NewElement(convertBE16x16ToLE26x10(api, c.SHi, c.SLo))
	// SUCCESS_BIT indicates if the input is a valid signature (1 for valid, 0
	// for invalid). Recall that we also allow to verify invalid signatures (for
	// the ECRECOVER precompile call).
	isFailure = api.Sub(1, c.SUCCESS_BIT)
	// ECRECOVERBIT indicates if the input comes from the ECRECOVER precompile
	// or not (1 for ECRECOVER, 0 for TX).
	strictRange = api.Sub(1, c.ECRECOVERBIT)

	return
}

var (
	plonkInputFillerKey = "ecdsa-secp256k1-plonk-input-filler"
)

func init() {
	plonk.RegisterInputFiller(plonkInputFillerKey, PlonkInputFiller)
}

// PlonkInputFiller is the input-filler that we use to assign the public inputs
// of incomplete circuits. This function must be registered via the
// [plonk.RegisterInputFiller] via the [init] function. But this has to be done
// manually if the package is not imported.
func PlonkInputFiller(circuitInstance, inputIndex int) field.Element {
	// every instance has 14 inputs.
	// pubkey xHi, pubkey xLo, pubkey yHi, pubkey yLo, hHi, hLo, vHi, vLo, rHi, rLo, sHi, sLo, successBit, ecrecoverBit

	// public key 1*G
	placeholderPubkey := [64]byte{0x79, 0xbe, 0x66, 0x7e, 0xf9, 0xdc, 0xbb, 0xac, 0x55, 0xa0, 0x62, 0x95, 0xce, 0x87, 0xb, 0x7, 0x2, 0x9b, 0xfc, 0xdb, 0x2d, 0xce, 0x28, 0xd9, 0x59, 0xf2, 0x81, 0x5b, 0x16, 0xf8, 0x17, 0x98, 0x48, 0x3a, 0xda, 0x77, 0x26, 0xa3, 0xc4, 0x65, 0x5d, 0xa4, 0xfb, 0xfc, 0xe, 0x11, 0x8, 0xa8, 0xfd, 0x17, 0xb4, 0x48, 0xa6, 0x85, 0x54, 0x19, 0x9c, 0x47, 0xd0, 0x8f, 0xfb, 0x10, 0xd4, 0xb8}
	// 1000000...0000
	placeholderTxHash := [32]byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	// valid signature for secret key 1 and tx hash 1 with random nonce
	// valid signature for secret key 1 and tx hash 1 with nonce 2
	placeholderSignature := [66]byte{
		// r part of signature
		0xc6, 0x4, 0x7f, 0x94, 0x41, 0xed, 0x7d, 0x6d, 0x30, 0x45, 0x40, 0x6e, 0x95, 0xc0, 0x7c, 0xd8, 0x5c, 0x77, 0x8e, 0x4b, 0x8c, 0xef, 0x3c, 0xa7, 0xab, 0xac, 0x9, 0xb9, 0x5c, 0x70, 0x9e, 0xe5,
		// s part of signature
		0xe3, 0x82, 0x3f, 0xca, 0x20, 0xf6, 0xbe, 0xb6, 0x98, 0x22, 0xa0, 0x37, 0x4a, 0xe0, 0x3e, 0x6b, 0x8b, 0x93, 0x35, 0x99, 0x1e, 0x1b, 0xee, 0x71, 0xb5, 0xbf, 0x34, 0x23, 0x16, 0x53, 0x70, 0x13,
		// v part of signature (in EVM format, 27 is added to the recovery id)
		0x0, 0x1b,
	}
	var ret field.Element
	switch inputIndex % 14 {
	case 0: // PK X HI
		ret.SetBytes(placeholderPubkey[0:16])
	case 1: // PK X LO
		ret.SetBytes(placeholderPubkey[16:32])
	case 2: // PK Y HI
		ret.SetBytes(placeholderPubkey[32:48])
	case 3: // PK Y LO
		ret.SetBytes(placeholderPubkey[48:64])
	case 4: // MSG HI
		ret.SetBytes(placeholderTxHash[0:16])
	case 5: // MSH LO
		ret.SetBytes(placeholderTxHash[16:32])
	case 6: // v HI
		ret.SetUint64(0)
	case 7: // v LO
		ret.SetUint64(0x1b)
	case 8: // r HI
		ret.SetBytes(placeholderSignature[0:16])
	case 9: // r LO
		ret.SetBytes(placeholderSignature[16:32])
	case 10: // s HI
		ret.SetBytes(placeholderSignature[32:48])
	case 11: // s LO
		ret.SetBytes(placeholderSignature[48:64])
	case 12: // success bit
		ret.SetUint64(1)
	case 13: // ecrecover bit
		ret.SetUint64(1)
	}
	return ret
}
