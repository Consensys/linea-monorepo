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

package ringsis

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/maths/common/poly"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

type MulModTest struct {
	P    [8]fr.Element
	Q, R [8]frontend.Variable
}

func (circuit *MulModTest) Define(api frontend.API) error {
	r := gnarkMulMod(api, circuit.P[:], circuit.Q[:])

	for i := 0; i < len(r); i++ {
		api.AssertIsEqual(r[i], circuit.R[i])
	}

	return nil
}

func TestMulMod(t *testing.T) {

	// get correct data
	p := make([]fr.Element, 8)
	q := make([]fr.Element, 8)

	for i := 0; i < 8; i++ {
		p[i].SetRandom()
		q[i].SetRandom()
	}

	r := mulMod(p, q)

	var witness MulModTest
	for i := 0; i < len(p); i++ {
		witness.P[i] = p[i]
		witness.Q[i] = q[i]
		witness.R[i] = r[i]
	}

	var circuit MulModTest
	for i := 0; i < len(circuit.P); i++ {
		circuit.P[i] = p[i]
	}
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit, frontend.IgnoreUnconstrainedInputs())
	if err != nil {
		t.Fatal(err)
	}

	twitness, err := frontend.NewWitness(&witness, ecc.BN254.ScalarField())
	if err != nil {
		t.Fatal(err)
	}
	err = ccs.IsSolved(twitness)
	if err != nil {
		t.Fatal(err)
	}

}

// SumTest
type SumTest struct {
	// Sis instance from gnark-crypto
	Sis Key
	// message to hash
	M []frontend.Variable
	// Expected result
	R []frontend.Variable
}

func (circuit *SumTest) Define(api frontend.API) error {
	// hash M
	h := circuit.Sis.GnarkHash(api, circuit.M)

	// check against the result
	for i := 0; i < len(h); i++ {
		api.AssertIsEqual(h[i], circuit.R[i])
	}
	return nil
}

// test 1: without padding
func TestSum1(t *testing.T) {

	// generate the witness
	// Sis with:
	// * key of size 8
	// * on \mathbb{Z}_r[X]/X^{8}+1
	// * with coefficients of M < 2^4 = 16
	// Note: this allows to hash 256bits to 256 bytes, so it's completely pointless
	// whith those parameters, it's for testing only
	key := (&Params{LogTwoBound_: 4, LogTwoDegree: 4}).GenerateKey(1)

	var toSum fr.Element
	toSum.SetString("5237501451071880303487629517413837912210424399515269294611144167440988308494")
	toSumBytes := toSum.Marshal()
	key.Write(toSumBytes)
	sum := key.Sum(nil)

	res := make([]fr.Element, key.Degree)
	for i := range res {
		res[i].SetBytes(sum[i*32 : (i+1)*32])
	}

	// witness
	var witness SumTest
	witness.M = make([]frontend.Variable, 1)
	witness.M[0] = toSum
	witness.R = make([]frontend.Variable, key.Degree)
	for i := range witness.R {
		witness.R[i] = res[i]
	}
	witness.Sis = key

	// circuit
	var circuit SumTest
	circuit.M = make([]frontend.Variable, 1)
	circuit.R = make([]frontend.Variable, key.Degree)
	circuit.Sis = key

	// compile...
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		t.Fatal(err)
	}

	// solve the circuit
	twitness, err := frontend.NewWitness(&witness, ecc.BN254.ScalarField())
	if err != nil {
		t.Fatal(err)
	}
	err = ccs.IsSolved(twitness)
	if err != nil {
		t.Fatal(err)
	}
}

// test 2: with padding
func TestSum2(t *testing.T) {

	key := (&Params{LogTwoBound_: 4, LogTwoDegree: 3}).GenerateKey(32)

	// data to hash, the buffer in the hasher will be padded, because
	// the correct size if [4]fr.Element.
	var toSum [3]fr.Element
	toSum[0].SetRandom()
	toSum[1].SetRandom()
	toSum[2].SetRandom()
	key.Write(toSum[0].Marshal())
	key.Write(toSum[1].Marshal())
	key.Write(toSum[2].Marshal())
	sum := key.Sum(nil)

	var res [8]fr.Element
	for i := 0; i < 8; i++ {
		res[i].SetBytes(sum[i*32 : (i+1)*32])
	}

	// witness
	var witness SumTest
	witness.M = make([]frontend.Variable, 3)
	witness.M[0] = toSum[0]
	witness.M[1] = toSum[1]
	witness.M[2] = toSum[2]
	witness.R = make([]frontend.Variable, 8)
	for i := 0; i < 8; i++ {
		witness.R[i] = res[i]
	}
	witness.Sis = key

	// circuit
	var circuit SumTest
	circuit.M = make([]frontend.Variable, 3)
	circuit.R = make([]frontend.Variable, 8)
	circuit.Sis = key

	// compile...
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		t.Fatal(err)
	}

	// solve the circuit
	twitness, err := frontend.NewWitness(&witness, ecc.BN254.ScalarField())
	if err != nil {
		t.Fatal(err)
	}
	err = ccs.IsSolved(twitness)
	if err != nil {
		t.Fatal(err)
	}
}

// Compute the polynomial product of p and q for testing. Assume p and q are both of length n
// Returns the result in the form (mod X^n + 1)
func mulMod(p, q []fr.Element) []fr.Element {
	// defer the polynomial evaluation routine
	res := poly.Mul(p, q)

	// performs the modular reduction
	for i := len(p); i < len(res); i++ {
		res[i-len(p)].Sub(&res[i-len(p)], &res[i])
	}

	return res
}
