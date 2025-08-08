package vortex

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/stretchr/testify/require"
)

// Evaluating a polynomial or its LDE yields the same result
func TestReedSolomonExtDoesNotChangeEvaluation(t *testing.T) {

	polySize := 1 << 10
	_nPolys := 15
	_blowUpFactor := 2

	x := fext.RandomElement()

	params := NewParams(_blowUpFactor, polySize, _nPolys, ringsis.StdParams, poseidon2.NewMerkleDamgardHasher, poseidon2.NewMerkleDamgardHasher)
	vec := smartvectors.RandExt(1 << 10)
	rsEncoded := params.rsEncodeExt(vec, nil)

	err := params.isCodewordExt(rsEncoded)
	require.NoError(t, err)

	y0 := smartvectors.EvaluateLagrangeFullFext(vec, x)
	y1 := smartvectors.EvaluateLagrangeFullFext(rsEncoded, x)

	require.Equal(t, y0.String(), y1.String())
}

// Evaluating and testing for constants
func TestReedSolomonExtConstant(t *testing.T) {

	polySize := 1 << 10
	_nPolys := 15
	_blowUpFactor := 2

	x := fext.RandomElement()

	params := NewParams(_blowUpFactor, polySize, _nPolys, ringsis.StdParams, poseidon2.NewMerkleDamgardHasher, poseidon2.NewMerkleDamgardHasher)
	vec := smartvectors.NewConstantExt(fext.RandomElement(), polySize)
	rsEncoded := params.rsEncodeExt(vec, nil)

	err := params.isCodewordExt(rsEncoded)
	require.NoError(t, err)

	y0 := smartvectors.EvaluateLagrangeFullFext(vec, x)
	y1 := smartvectors.EvaluateLagrangeFullFext(rsEncoded, x)

	require.Equal(t, y0.String(), y1.String())

}
