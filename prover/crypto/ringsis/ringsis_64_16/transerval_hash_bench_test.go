package ringsis_64_16_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis/ringsis_64_16"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
)

func BenchmarkTransversalHash(b *testing.B) {

	var (
		numRow          = 1024
		numCols         = 1024
		rng             = rand.New(utils.NewRandSource(786868)) // nolint
		domain          *fft.Domain
		twiddles        = ringsis_64_16.PrecomputeTwiddlesCoset(domain.Generator, domain.FrMultiplicativeGen)
		params          = ringsis.Params{LogTwoBound: 16, LogTwoDegree: 6}
		numInputPerPoly = params.OutputSize() / (field.Bytes * 8 / params.LogTwoBound)
		key             = ringsis.GenerateKey(params, numRow)
		numTestCases    = 1 << numInputPerPoly
		numPoly         = numRow / numInputPerPoly
	)
	omega, err := fft.Generator(64 * 2)
	if err != nil {
		panic(err)
	}
	domain = fft.NewDomain(64, fft.WithShift(omega))

	for tc := 0; tc < numTestCases; tc++ {

		b.Run(fmt.Sprintf("testcase-%b", tc), func(b *testing.B) {

			inputs := make([]smartvectors.SmartVector, 0, numPoly*numInputPerPoly)

			for p := 0; p < numPoly; p++ {
				for i := 0; i < numInputPerPoly; i++ {
					if (tc>>i)&1 == 0 {
						inputs = append(inputs, randomConstRow(rng, numCols))
					} else {
						inputs = append(inputs, randomRegularRow(rng, numCols))
					}
				}
			}

			b.ResetTimer()

			for c := 0; c < b.N; c++ {
				_ = ringsis_64_16.TransversalHash(key.Ag(), inputs, twiddles, domain)
			}

		})

	}
}
