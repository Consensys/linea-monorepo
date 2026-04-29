package plonk2

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNTTPlan_OrderContracts(t *testing.T) {
	for _, tc := range []struct {
		name      string
		direction nttDirection
		in        nttOrder
		out       nttOrder
	}{
		{name: "forward", direction: nttDirectionForward, in: nttOrderNatural, out: nttOrderBitReversed},
		{name: "inverse", direction: nttDirectionInverse, in: nttOrderBitReversed, out: nttOrderNatural},
		{name: "coset-forward", direction: nttDirectionCosetForward, in: nttOrderNatural, out: nttOrderNatural},
		{name: "coset-inverse", direction: nttDirectionCosetInverse, in: nttOrderNatural, out: nttOrderNatural},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := defaultNTTPlan(CurveBLS12377, 1024, tc.direction, 3)
			require.NoError(t, err)
			require.Equal(t, tc.in, plan.InputOrder, "input order should match current FFT contract")
			require.Equal(t, tc.out, plan.OutputOrder, "output order should match current FFT contract")
			require.Equal(t, nttResidencyDevice, plan.InputResidency, "current API consumes device vectors")
			require.Equal(t, nttResidencyDevice, plan.OutputResidency, "current API produces device vectors")
			require.Equal(t, 3, plan.BatchCount, "batch count should be retained")
		})
	}
}

func TestNTTPlan_TargetCurves(t *testing.T) {
	for _, curve := range []Curve{CurveBN254, CurveBLS12377, CurveBW6761} {
		plan, err := defaultNTTPlan(curve, 1<<16, nttDirectionForward, 1)
		require.NoError(t, err)
		require.Equal(t, curve, plan.Curve)
		require.Equal(t, 1<<16, plan.Size)
	}
}

func TestNTTPlan_InvalidInputs(t *testing.T) {
	_, err := defaultNTTPlan(CurveBN254, 1000, nttDirectionForward, 1)
	require.Error(t, err, "non-power-of-two size should fail")

	_, err = defaultNTTPlan(CurveBN254, 1024, nttDirectionForward, 0)
	require.Error(t, err, "empty batch should fail")

	_, err = defaultNTTPlan(CurveBN254, 1024, nttDirection(99), 1)
	require.Error(t, err, "unknown direction should fail")
}
