//go:build !fuzzlight

package gkrposeidon2_test

import (
	"fmt"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	poseidon2native "github.com/consensys/linea-monorepo/prover/crypto/poseidon2_bls12377"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_bls12377/gkrposeidon2"
	"github.com/stretchr/testify/require"
)

type SimpleHashingCircuit struct {
	Input  []frontend.Variable
	Digest frontend.Variable
}

func (c SimpleHashingCircuit) Define(api frontend.API) error {
	factory := gkrposeidon2.NewHasherFactory(api)
	h := factory.NewCompresser()
	h.Write(c.Input...)
	api.AssertIsEqual(h.Sum(), c.Digest)
	return nil
}

func TestFactory(t *testing.T) {
	ccs, err := frontend.Compile(
		ecc.BLS12_377.ScalarField(),
		scs.NewBuilder,
		&SimpleHashingCircuit{Input: make([]frontend.Variable, 4)},
	)
	require.NoError(t, err)

	vals := []fr.Element{fr.NewElement(0), fr.NewElement(1), fr.NewElement(2), fr.NewElement(3)}
	hNative := poseidon2native.NewMDHasher()
	hNative.WriteElements(vals...)
	digest := hNative.SumElement()

	assignment := SimpleHashingCircuit{
		Input:  []frontend.Variable{vals[0].String(), vals[1].String(), vals[2].String(), vals[3].String()},
		Digest: digest.String(),
	}

	witness, err := frontend.NewWitness(&assignment, ecc.BLS12_377.ScalarField())
	require.NoError(t, err)
	require.NoError(t, ccs.IsSolved(witness))
}

func TestFactoryManySizes(t *testing.T) {
	for _, size := range []int{1, 2, 3, 12, 37, 64, 110, 128, 129, 221, 240, 256, 260} {
		t.Run(fmt.Sprintf("%d-hashes", size), func(t *testing.T) {
			ccs, err := frontend.Compile(
				ecc.BLS12_377.ScalarField(),
				scs.NewBuilder,
				&SimpleHashingCircuit{Input: make([]frontend.Variable, size)},
			)
			require.NoError(t, err)

			vals := make([]fr.Element, size)
			inputs := make([]frontend.Variable, size)
			for i := range inputs {
				vals[i] = fr.NewElement(uint64(10 + i))
				inputs[i] = vals[i].String()
			}

			hNative := poseidon2native.NewMDHasher()
			hNative.WriteElements(vals...)
			digest := hNative.SumElement()

			assignment := SimpleHashingCircuit{
				Input:  inputs,
				Digest: digest.String(),
			}
			witness, err := frontend.NewWitness(&assignment, ecc.BLS12_377.ScalarField())
			require.NoError(t, err)
			require.NoError(t, ccs.IsSolved(witness))
		})
	}
}
