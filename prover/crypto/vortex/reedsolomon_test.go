package vortex

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/stretchr/testify/require"
)

// Evaluating a polynomial or its LDE yields the same result
func TestReedSolomonDoesNotChangeEvaluation(t *testing.T) {

	polySize := 1 << 10
	_nPolys := 15
	_blowUpFactor := 2

	x := fext.RandomElement()

	params := NewParams(_blowUpFactor, polySize, _nPolys, ringsis.StdParams, poseidon2.NewMerkleDamgardHasher, nil)
	vec := smartvectors.Rand(1 << 10)
	rsEncoded := params._rsEncodeBase(vec)

	err := params.isCodeword(rsEncoded)
	require.NoError(t, err)

	y0 := smartvectors.EvaluateBasePolyLagrange(vec, x)
	y1 := smartvectors.EvaluateBasePolyLagrange(rsEncoded, x)

	require.Equal(t, y0.String(), y1.String())
}

// Evaluating and testing for constants
func TestReedSolomonConstant(t *testing.T) {

	polySize := 1 << 10
	_nPolys := 15
	_blowUpFactor := 2

	x := fext.RandomElement()

	params := NewParams(_blowUpFactor, polySize, _nPolys, ringsis.StdParams, poseidon2.NewMerkleDamgardHasher, nil)
	vec := smartvectors.NewConstant(field.NewElement(42), polySize)
	rsEncoded := params._rsEncodeBase(vec)

	err := params.isCodeword(rsEncoded)
	require.NoError(t, err)

	y0 := smartvectors.EvaluateBasePolyLagrange(vec, x)
	y1 := smartvectors.EvaluateBasePolyLagrange(rsEncoded, x)

	require.Equal(t, y0.String(), y1.String())

}
