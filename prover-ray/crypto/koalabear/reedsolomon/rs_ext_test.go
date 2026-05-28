// Copyright Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
// specific language governing permissions and limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package reedsolomon

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	ext "github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover-ray/crypto/koalabear/poly"
)

func e4FromU64(a0, a1, b0, b1 uint64) ext.E4 {
	var z ext.E4
	z.B0.A0.SetUint64(a0)
	z.B0.A1.SetUint64(a1)
	z.B1.A0.SetUint64(b0)
	z.B1.A1.SetUint64(b1)
	return z
}

func liftE4(v koalabear.Element) ext.E4 {
	var z ext.E4
	z.Lift(&v)
	return z
}

func canonicalEvalExt(coeffs []ext.E4, z ext.E4) ext.E4 {
	if len(coeffs) == 0 {
		return ext.E4{}
	}
	y := coeffs[len(coeffs)-1]
	for i := len(coeffs) - 2; i >= 0; i-- {
		y.Mul(&y, &z)
		y.Add(&y, &coeffs[i])
	}
	return y
}

func TestEncodeExt(t *testing.T) {
	coeffs := poly.ExtPolynomial{
		e4FromU64(1, 2, 3, 4),
		e4FromU64(5, 6, 7, 8),
		e4FromU64(9, 10, 11, 12),
		e4FromU64(13, 14, 15, 16),
	}

	domainD := fft.NewDomain(uint64(len(coeffs)))
	p := make(poly.ExtPolynomial, len(coeffs))
	copy(p, coeffs)
	domainD.FFTExt(p, fft.DIF)
	fft.BitReverse(p)

	encoder := NewEncoder(8)
	encoded := encoder.EncodeExt(p, domainD)

	domainN := fft.NewDomain(uint64(len(encoded)))
	omega := domainN.Generator
	var omegaJ koalabear.Element
	omegaJ.SetOne()
	for j := range encoded {
		x := liftE4(omegaJ)
		want := canonicalEvalExt(coeffs, x)
		if !encoded[j].Equal(&want) {
			t.Fatalf("encoded[%d] = %s, want %s", j, encoded[j].String(), want.String())
		}
		omegaJ.Mul(&omegaJ, &omega)
	}
}
