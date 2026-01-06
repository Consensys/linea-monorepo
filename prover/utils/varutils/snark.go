package varutils

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/compress"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/gnark/std/rangecheck"
	"github.com/consensys/linea-monorepo/prover/maths/zk"
)

func Copy[T any](dst []frontend.Variable, src []T) (n int) {
	n = min(len(dst), len(src))

	for i := 0; i < n; i++ {
		dst[i] = zk.ValueOf(src[i])
	}

	return
}

// ToBytes32 decomposes x into 32 bytes.
func ToBytes32(api frontend.API, x frontend.Variable) [32]frontend.Variable {
	var res [32]frontend.Variable
	d, err := ToBytes(api, []frontend.Variable{x}, fr.Bytes*8)
	if err != nil {
		panic(err)
	}
	slack := 32 - len(d) // should be zero
	copy(res[slack:], d)
	for i := range slack {
		res[i] = 0
	}
	return res
}

func RegisterHints() {
	solver.RegisterHint(breakIntoBytesHint, divBy31Hint)
}

// ReduceBytes reduces given bytes modulo a given field. As a side effect, the "bytes" are range checked
func ReduceBytes[T emulated.FieldParams](api frontend.API, bytes []*frontend.Variable) []frontend.Variable {

	bits := api.ToBinary(NewElementFromBytes[T](api, bytes))
	res := make([]frontend.Variable, (len(bits)+7)/8)
	copy(bits[:], bits)
	for i := len(bits); i < len(bits); i++ {
		bits[i] = 0
	}
	for i := range res {
		bitsStart := 8 * (len(res) - i - 1)
		bitsEnd := bitsStart + 8
		if i == 0 {
			bitsEnd = len(bits)
		}
		res[i] = api.FromBinary(bits[bitsStart:bitsEnd]...)
	}

	return res
}

// NewElementFromBytes range checks the bytes and gives a reduced field element
func NewElementFromBytes[T emulated.FieldParams](api frontend.API, bytes []*frontend.Variable) frontend.Variable {

	bits := make([]frontend.Variable, 8*len(bytes))
	for i := range bytes {
		copy(bits[8*i:], api.ToBinary(bytes[len(bytes)-i-1], 8))
	}

	f, err := emulated.NewField[T](api)
	if err != nil {
		panic(err)
	}

	return f.Reduce(f.Add(f.FromBits(bits...), f.Zero()))

	return nil
}

func ToVariableSlice[X any](s []X) []frontend.Variable {
	res := make([]frontend.Variable, len(s))
	Copy(res, s)
	return res
}

// ToBytes takes data words each containing wordNbBits many bits, and repacks them as 8-bit bytes.
// The data will be range checked only if the words are larger than a byte.
func ToBytes(api frontend.API, data []frontend.Variable, wordNbBits int) ([]frontend.Variable, error) {
	if wordNbBits == 8 {
		return data, nil
	}

	if wordNbBits < 8 {
		wordsPerByte := 8 / wordNbBits
		if wordsPerByte*wordNbBits != 8 {
			return nil, fmt.Errorf("currently only multiples or quotients of bytes supported, not the case for the given %d-bit words", wordNbBits)
		}
		radix := big.NewInt(1 << uint(wordNbBits))
		bytes := make([]frontend.Variable, len(data)*wordNbBits/8)
		for i := range bytes {
			bytes[i] = compress.ReadNum(api, data[i*wordsPerByte:i*wordsPerByte+wordsPerByte], radix)
		}
		return bytes, nil
	}

	bytesPerWord := wordNbBits / 8
	if bytesPerWord*8 != wordNbBits {
		return nil, fmt.Errorf("currently only multiples or quotients of bytes supported, not the case for the given %d-bit words", wordNbBits)
	}
	bytes, err := api.Compiler().NewHint(breakIntoBytesHint, len(data)*wordNbBits/8, data...)
	if err != nil {
		return nil, err
	}
	rc := rangecheck.New(api)
	for _, b := range bytes {
		rc.Check(b, 8)
	}

	radix := big.NewInt(256)
	for i := range data {
		api.AssertIsEqual(data[i], compress.ReadNum(api, bytes[i*bytesPerWord:i*bytesPerWord+bytesPerWord], radix))
	}

	return bytes, nil
}

func breakIntoBytesHint(_ *big.Int, words []*big.Int, bytes []*big.Int) error {
	bytesPerWord := len(bytes) / len(words)
	if bytesPerWord*len(words) != len(bytes) {
		return errors.New("words are not byte aligned")
	}

	for i := range words {
		b := words[i].Bytes()
		if len(b) > bytesPerWord {
			return fmt.Errorf("word #%d doesn't fit in %d bytes: 0x%s", i, bytesPerWord, words[i].Text(16))
		}
		for j := range b {
			bytes[i*bytesPerWord+j+bytesPerWord-len(b)].SetUint64(uint64(b[j]))
		}
	}
	return nil
}

// DivBy31 returns q, r such that v = 31 q + r, and 0 ≤ r < 31
// side effect: ensures 0 ≤ v[i] < 2ᵇⁱᵗˢ⁺².
func DivBy31(api frontend.API, v frontend.Variable, bits int) (q, r frontend.Variable, err error) {
	_q, _r, err := DivManyBy31(api, []frontend.Variable{v}, bits)
	if err != nil {
		return nil, nil, err
	}
	return _q[0], _r[0], nil
}

// DivManyBy31 returns q, r for each v such that v = 31 q + r, and 0 ≤ r < 31
// side effect: ensures 0 ≤ v[i] < 2ᵇⁱᵗˢ⁺² for all i
func DivManyBy31(api frontend.API, v []frontend.Variable, bits int) (q, r []frontend.Variable, err error) {
	qNbBits := bits - 4

	if hintOut, err := api.Compiler().NewHint(divBy31Hint, 2*len(v), v...); err != nil {
		return nil, nil, err
	} else {
		q, r = hintOut[:len(v)], hintOut[len(v):]
	}

	rChecker := rangecheck.New(api)

	for i := range v { // TODO See if lookups or api.AssertIsLte would be more efficient
		rChecker.Check(r[i], 5)
		api.AssertIsDifferent(r[i], 31)
		rChecker.Check(q[i], qNbBits)
		api.AssertIsEqual(v[i], api.Add(api.Mul(q[i], 31), r[i])) // 31 × q < 2ᵇⁱᵗˢ⁻⁴ 2⁵ ⇒ v < 2ᵇⁱᵗˢ⁺¹ + 31 < 2ᵇⁱᵗˢ⁺²
	}
	return q, r, nil
}

// outs: [quotients], [remainders]
func divBy31Hint(_ *big.Int, ins []*big.Int, outs []*big.Int) error {
	if len(outs) != 2*len(ins) {
		return errors.New("expected output layout: [quotients][remainders]")
	}

	q := outs[:len(ins)]
	r := outs[len(ins):]
	for i := range ins {
		v := ins[i].Uint64()
		q[i].SetUint64(v / 31)
		r[i].SetUint64(v % 31)
	}

	return nil
}
