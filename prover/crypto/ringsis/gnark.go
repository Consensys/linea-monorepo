package ringsis

// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import (
	"errors"

	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/frontend"
)

var (
	ErrWrongSize = errors.New("size does not fit")
)

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the underlying hash state.
// The instance buffer is interpreted as a sequence of coefficients of size r.Bound bits long.
// The function returns the hash of the polynomial as a a sequence []fr.Elements, interpreted as []bytes,
// corresponding to sum_i A[i]*m Mod X^{d}+1
func (r *Key) GnarkHash(api frontend.API, b []frontend.Variable) []frontend.Variable {

	if len(b) > r.MaxNbFieldToHash {
		panic("buffer too large")
	}

	numBitsPerField := r.NumLimbs() * r.LogTwoBound
	numBitsToSum := len(b) * numBitsPerField
	vBits := make([]frontend.Variable, numBitsToSum)

	// Unpack the bits of the construction
	for i := range b {
		// The reverse the result of api.ToBinary, accounting for the fact that
		// it returns the bits in little-endian while gnark's version of the software.
		littleEndian := api.ToBinary(b[i], field.Bits)
		littleEndian = append(littleEndian, 0, 0) // Pad to 256 with constants

		copy(vBits[i*numBitsPerField:(i+1)*numBitsPerField], littleEndian)
	}

	// we process the input buffer by blocks of r.LogTwoBound bits
	// each of these block (<< 64bits) are interpreted as a coefficient
	nbCoefficients := r.NumLimbs() * len(b)
	m := make([]frontend.Variable, nbCoefficients)
	for i := range m {
		m[i] = api.FromBinary(vBits[i*r.LogTwoBound : (i+1)*r.LogTwoBound]...)
	}

	if len(m)%r.Degree != 0 {
		nbToAdd := r.Degree - (len(m) % r.Degree)
		for i := 0; i < nbToAdd; i++ {
			m = append(m, 0)
		}
	}

	// preallocated the result
	res := make([]frontend.Variable, r.Degree)
	for i := range res {
		res[i] = 0
	}

	for i := 0; i*r.Degree < nbCoefficients; i++ {
		k := m[i*r.Degree : (i+1)*r.Degree]
		var tmp []frontend.Variable

		// Account for the fact that the underlying ringSIS instance
		// does not actually perform the Montgommery form conversion
		// this equivalent to multiplying the SIS key by the inverse
		// of the montgommery constant.
		aRInv := make([]field.Element, r.Degree)
		for j := range aRInv {
			aRInv[j] = MulRInv(r.A[i][j])
		}

		switch r.Degree {
		case 2:
			tmp = gnarkMulModDegree2(api, aRInv, k)
		default:
			tmp = gnarkMulMod(api, aRInv, k)
		}

		for i := range res {
			res[i] = api.Add(res[i], tmp[i])
		}
	}

	return res
}

func gnarkMulModDegree2(api frontend.API, p []fr.Element, q []frontend.Variable) []frontend.Variable {
	// (p0+p1*X)*(q0+q1*X) Mod X^2+1 = p0q0-p1q1+(p0q1+p1q0)*X
	// We do that in 3 muls instead of 4:
	// a = p0q0
	// b = p1q1
	// c = (p0+p1)*(q0+q1)
	// r = a - b + (c-a-b)*X
	var res [2]frontend.Variable
	var a, b, c frontend.Variable
	a = api.Mul(p[0], q[0])
	b = api.Mul(p[1], q[1])
	res[0] = api.Sub(a, b)
	c = api.Add(p[0], p[1])
	res[1] = api.Add(q[0], q[1])
	res[1] = api.Mul(res[1], c)
	res[1] = api.Sub(res[1], a)
	res[1] = api.Sub(res[1], b)
	return res[:]
}

// mulMod computes p * q Mod X^d+1 where d = len(p) = len(q).
// It is assumed that p and q are of the same size.
func gnarkMulMod(api frontend.API, p []fr.Element, q []frontend.Variable) []frontend.Variable {

	d := len(p)
	res := make([]frontend.Variable, d)
	for i := 0; i < d; i++ {
		res[i] = 0
	}

	for i := 0; i < d; i++ {
		for j := 0; j < d-i; j++ {
			res[i+j] = api.Add(api.Mul(p[j], q[i]), res[i+j])
		}
		for j := d - i; j < d; j++ {
			res[j-d+i] = api.Sub(res[j-d+i], api.Mul(p[j], q[i]))
		}
	}

	return res
}

var f field.Element = func() field.Element {
	var x field.Element
	x.SetString("9915499612839321149637521777990102151350674507940716049588462388200839649614")
	return x
}()

// Multiply the field element by R^-1, where R is the Montgommery constant
func MulRInv(x field.Element) field.Element {
	var res field.Element
	res.Mul(&x, &f)
	return res
}

var r field.Element = func() field.Element {
	var x field.Element
	x.SetString("6350874878119819312338956282401532410528162663560392320966563075034087161851")
	return x
}()

// Multiply by R, where R is the Montgommery constant
func MulR(x field.Element) field.Element {
	var res field.Element
	res.Mul(&x, &r)
	return res
}
