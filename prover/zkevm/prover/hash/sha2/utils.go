package sha2

import (
	"errors"
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/rangecheck"
)

// Decompose x in 'nBytes' bytes in big endian order
//
// Deprecated: These are utility functions that have been copy-pasted from circuits/internal
// waiting for them or equivalent function to be merged in gnark/std. We will
// be able to substitute them at this point.
func toNBytes(api frontend.API, x frontend.Variable, nBytes int) []frontend.Variable {
	return decomposeIntoBytes(api, x, nBytes)
}

func decomposeIntoBytes(api frontend.API, data frontend.Variable, nbBytes int) []frontend.Variable {

	bytes, err := api.Compiler().NewHint(decomposeIntoBytesHint, nbBytes, data)
	if err != nil {
		panic(err)
	}

	var (
		rc     = rangecheck.New(api)
		recmpt = frontend.Variable(0)
	)

	for i := 0; i < nbBytes; i++ {
		rc.Check(bytes[i], 8)
		recmpt = api.Mul(recmpt, 256)
		recmpt = api.Add(recmpt, bytes[i])
	}

	api.AssertIsEqual(recmpt, data)

	return bytes
}

func decomposeIntoBytesHint(_ *big.Int, ins, outs []*big.Int) error {
	nbBytes := len(outs) / len(ins)
	if nbBytes*len(ins) != len(outs) {
		return errors.New("incongruent number of ins/outs")
	}
	var v, radix, zero big.Int
	radix.SetUint64(256)
	for i := range ins {
		v.Set(ins[i])
		for j := nbBytes - 1; j >= 0; j-- {
			outs[i*nbBytes+j].Mod(&v, &radix)
			v.Rsh(&v, 8)
		}
		if v.Cmp(&zero) != 0 {
			return errors.New("not fitting in len(outs)/len(ins) many bytes")
		}
	}
	return nil
}
