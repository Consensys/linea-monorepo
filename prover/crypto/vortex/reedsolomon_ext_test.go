package vortex

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
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

	params := NewParams(_blowUpFactor, polySize, _nPolys, ringsis.StdParams, poseidon2_koalabear.Poseidon2, poseidon2_koalabear.Poseidon2)
	vec := smartvectors.RandExt(1 << 10)
	rsEncoded := params.rsEncodeExt(vec)

	err := params.IsCodewordExt(rsEncoded)
	require.NoError(t, err)

	y0 := smartvectors.EvaluateFextPolyLagrange(vec, x)
	y1 := smartvectors.EvaluateFextPolyLagrange(rsEncoded, x)

	require.Equal(t, y0.String(), y1.String())
}

// Evaluating and testing for constants
func TestReedSolomonExtConstant(t *testing.T) {

	polySize := 1 << 10
	_nPolys := 15
	_blowUpFactor := 2

	x := fext.RandomElement()

	params := NewParams(_blowUpFactor, polySize, _nPolys, ringsis.StdParams, poseidon2_koalabear.Poseidon2, poseidon2_koalabear.Poseidon2)
	vec := smartvectors.NewConstantExt(fext.RandomElement(), polySize)
	rsEncoded := params.rsEncodeExt(vec)

	err := params.IsCodewordExt(rsEncoded)
	require.NoError(t, err)

	y0 := smartvectors.EvaluateFextPolyLagrange(vec, x)
	y1 := smartvectors.EvaluateFextPolyLagrange(rsEncoded, x)

	require.Equal(t, y0.String(), y1.String())

}
