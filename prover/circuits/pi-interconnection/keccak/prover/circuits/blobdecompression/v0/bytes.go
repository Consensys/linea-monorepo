package v0

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/utils"
	"github.com/sirupsen/logrus"
)

// Pack an array of bytes as if they were in big endian order into a single fe.
// returns 0 on an empty array and assumes the input bytes are already
// constrained into bytehood. If more bytes are provided than what can be
// contained in a single field element it overflows.
func packVarBytesBE(api frontend.API, bs []frontend.Variable) frontend.Variable {
	res := frontend.Variable(0)
	for i := range bs {
		if i > 0 {
			res = api.Mul(res, frontend.Variable(1<<8))
		}
		res = api.Add(res, bs[i])
	}
	return res
}

// Pack an array of variables (representing bytes) into variables capable of
// storing multiple bytes. The packing is done as tightly as possible.
func packVarByteSliceBE(api frontend.API, bs []frontend.Variable) []frontend.Variable {
	res := make([]frontend.Variable, utils.DivExact(len(bs), wordByteSize))
	for i := range res {
		res[i] = packVarBytesBE(api, bs[wordByteSize*i:wordByteSize*(i+1)])
	}
	return res
}

// Pack an array of limbs as if they were in big endian order into a single fe.
// returns 0 on an empty array and assumes the input bytes are already
// constrained into bytehood. If more bytes are provided than what can be
// contained in a single field element it overflows. The limbs are assumed to
// be 64 bytes.
func packVarLimbsBE(api frontend.API, bs []frontend.Variable) frontend.Variable {
	res := frontend.Variable(0)
	base := big.NewInt(1)
	base.Lsh(base, 64)
	for i := range bs {
		if i > 0 {
			res = api.Mul(res, base)
		}
		res = api.Add(res, bs[i])
	}
	return res
}

// Assigns a size-bounded array of bytes into an array of bytesVariable. The
// assigned bytes are zero-padded on the right.
func assignVarByteSlice(data []byte, maxLen int) ([]frontend.Variable, error) {

	if len(data) > maxLen {
		return nil, fmt.Errorf("bytesToVar : data too large : %v max is %v", len(data), maxLen)
	}

	res := make([]frontend.Variable, maxLen)

	for i := range res {
		if i < len(data) {
			res[i] = frontend.Variable(data[i])
		} else {
			res[i] = 0
		}
	}

	return res, nil
}

func packBytesInWords(data []byte, maxLen int) ([]fr.Element, error) {
	if len(data) > maxLen {
		logrus.Errorf("bytesToField : data too large : %v max is %v => truncating", len(data), maxLen)
	}
	maxNumWords := (maxLen + wordByteSize - 1) / wordByteSize
	res := make([]fr.Element, maxNumWords)

	for i := range res {
		chunk := [32]byte{}
		if wordByteSize*i < len(data) {
			copy(chunk[:], data[wordByteSize*i:])
		}
		if err := res[i].SetBytesCanonical(chunk[:]); err != nil {
			panic(err)
		}
	}

	return res, nil
}
