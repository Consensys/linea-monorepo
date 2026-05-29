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
	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/linea-monorepo/prover-ray/crypto/koalabear/poly"
)

func TestEvaluateOnExtendedDomainRootMatchesEncode(t *testing.T) {
	p := poly.Polynomial{
		rsBaseElement(1),
		rsBaseElement(2),
		rsBaseElement(3),
		rsBaseElement(4),
	}

	domainD := fft.NewDomain(uint64(len(p)))
	encoder := NewEncoder(8)
	encoded := encoder.Encode(p, domainD)

	for j := range encoded {
		got := poly.EvaluateOnExtendedDomainRoot(p, domainD, encoder.Domain, j)
		if !got.Equal(&encoded[j]) {
			t.Fatalf("evaluation[%d] = %s, want %s", j, got.String(), encoded[j].String())
		}
	}
}

func TestExtEvaluateOnExtendedDomainRootMatchesEncodeExt(t *testing.T) {
	p := poly.ExtPolynomial{
		e6FromU64(1, 2, 3, 4),
		e6FromU64(5, 6, 7, 8),
		e6FromU64(9, 10, 11, 12),
		e6FromU64(13, 14, 15, 16),
	}

	domainD := fft.NewDomain(uint64(len(p)))
	encoder := NewEncoder(8)
	encoded := encoder.EncodeExt(p, domainD)

	for j := range encoded {
		got := poly.ExtEvaluateOnExtendedDomainRoot(p, domainD, encoder.Domain, j)
		if !got.Equal(&encoded[j]) {
			t.Fatalf("evaluation[%d] = %s, want %s", j, got.String(), encoded[j].String())
		}
	}
}

func rsBaseElement(v uint64) koalabear.Element {
	var e koalabear.Element
	e.SetUint64(v)
	return e
}
