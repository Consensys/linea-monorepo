package smartvectors_mixed

import (
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear/fft"
	"github.com/consensys/gnark-crypto/field/koalabear/vortex"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/stretchr/testify/require"
)

func TestBatchEvaluateLagrange(t *testing.T) {

	x := fext.RandomElement()

	nbPoly := 8
	size := 64

	polys := make([][]field.Element, nbPoly)
	for i := 0; i < nbPoly; i++ {
		polys[i] = make([]field.Element, size)
		for j := 0; j < size; j++ {
			polys[i][j].SetRandom()
		}
	}

	evalCan := make([]fext.Element, nbPoly)
	for i := 0; i < nbPoly; i++ {
		evalCan[i] = vortex.EvalBasePolyHorner(polys[i], x)
	}

	d := fft.NewDomain(uint64(size))
	polyLagranges := make([][]field.Element, nbPoly)
	copy(polyLagranges, polys)

	polyLagrangeSv := make([]smartvectors.SmartVector, nbPoly)
	for i := 0; i < nbPoly; i++ {
		d.FFT(polyLagranges[i], fft.DIF)
		fft.BitReverse(polyLagranges[i])
		polyLagrangeSv[i] = smartvectors.NewRegular(polyLagranges[i])

	}
	evalLag := BatchEvaluateLagrange(polyLagrangeSv, x)

	// check the result
	for i := 0; i < nbPoly; i++ {
		require.Equal(t, evalLag[i].String(), evalCan[i].String())
	}

}
