//go:build skiptest

// this test is skipped in the test suite due to the fact that the evaluation benchmark
// files are outdated
package symbolic_test

import (
	"fmt"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"path"
	"testing"

	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/maths/common/mempool"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/stretchr/testify/assert"
)

func BenchmarkEvaluationSingleThreaded(b *testing.B) {

	makeRegularExt := func() smartvectors.SmartVector {
		return smartvectors.RandExt(symbolic.MaxChunkSize)
	}

	makeConstExt := func() smartvectors.SmartVector {
		var x fext.Element
		x.SetRandom()
		return smartvectors.NewConstantExt(x, symbolic.MaxChunkSize)
	}

	makeFullZeroExt := func() smartvectors.SmartVector {
		return smartvectors.NewConstantExt(fext.Zero(), symbolic.MaxChunkSize)
	}

	makeFullOnesExt := func() smartvectors.SmartVector {
		return smartvectors.NewConstantExt(fext.One(), symbolic.MaxChunkSize)
	}

	for ratio := 1; ratio <= 32; ratio *= 2 {

		b.Run(fmt.Sprintf("ratio-%v", ratio), func(b *testing.B) {

			var (
				testDir                           = "testdata/evaluation-benchmark"
				constanthoodFName                 = fmt.Sprintf("global-variable-constanthood-%v.csv", ratio)
				exprFName                         = fmt.Sprintf("global-cs-ratio-%v.cbor.gz", ratio)
				constanthoodFPath                 = path.Join(testDir, constanthoodFName)
				exprFPath                         = path.Join(testDir, exprFName)
				constantHoodFile                  = files.MustRead(constanthoodFPath)
				exprFile                          = files.MustReadCompressed(exprFPath)
				constantHood                      = symbolic.ReadConstanthoodFromCsv(constantHoodFile)
				expr                              = serialization.UnmarshalExprCBOR(exprFile)
				inputs                            = make([]smartvectors.SmartVector, len(constantHood))
				board                             = expr.Board()
				pool              mempool.MemPool = mempool.CreateFromSyncPool(symbolic.MaxChunkSize)
			)

			for i := range inputs {
				switch {
				case !constantHood[i][0]:
					inputs[i] = makeRegularExt()
				case constantHood[i][1]:
					inputs[i] = makeFullZeroExt()
				case constantHood[i][2]:
					inputs[i] = makeFullOnesExt()
				default:
					inputs[i] = makeConstExt()
				}
			}

			b.ResetTimer()

			for c := 0; c < b.N; c++ {
				_ = board.EvaluateExt(inputs, pool)
			}
		})
	}
}

func TestEvaluationSingleThreaded(t *testing.T) {

	makeRegularExt := func() smartvectors.SmartVector {
		return smartvectors.RandExt(symbolic.MaxChunkSize)
	}

	makeConstExt := func() smartvectors.SmartVector {
		var x fext.Element
		x.SetRandom()
		return smartvectors.NewConstantExt(x, symbolic.MaxChunkSize)
	}

	makeFullZeroExt := func() smartvectors.SmartVector {
		return smartvectors.NewConstantExt(fext.Zero(), symbolic.MaxChunkSize)
	}

	makeFullOnesExt := func() smartvectors.SmartVector {
		return smartvectors.NewConstantExt(fext.One(), symbolic.MaxChunkSize)
	}

	for ratio := 1; ratio <= 32; ratio *= 2 {

		t.Run(fmt.Sprintf("ratio-%v", ratio), func(b *testing.T) {

			var (
				testDir                           = "testdata/evaluation-benchmark"
				constanthoodFName                 = fmt.Sprintf("global-variable-constanthood-%v.csv", ratio)
				exprFName                         = fmt.Sprintf("global-cs-ratio-%v.cbor.gz", ratio)
				constanthoodFPath                 = path.Join(testDir, constanthoodFName)
				exprFPath                         = path.Join(testDir, exprFName)
				constantHoodFile                  = files.MustRead(constanthoodFPath)
				exprFile                          = files.MustReadCompressed(exprFPath)
				constantHood                      = symbolic.ReadConstanthoodFromCsv(constantHoodFile)
				expr                              = serialization.UnmarshalExprCBOR(exprFile)
				inputs                            = make([]smartvectors.SmartVector, len(constantHood))
				board                             = expr.Board()
				pool_             mempool.MemPool = mempool.CreateFromSyncPool(symbolic.MaxChunkSize)
			)

			pool_ = mempool.WrapsWithMemCache(pool_)
			pool := mempool.NewDebugPool(pool_)

			_, mustBeTrue := mempool.ExtractCheckOptionalSoft(symbolic.MaxChunkSize, pool)
			_, mustBeTrue2 := mempool.ExtractCheckOptionalStrict(symbolic.MaxChunkSize, pool)

			assert.True(t, mustBeTrue)
			assert.True(t, mustBeTrue2)

			for i := range inputs {
				switch {
				case !constantHood[i][0]:
					inputs[i] = makeRegularExt()
				case constantHood[i][1]:
					inputs[i] = makeFullZeroExt()
				case constantHood[i][2]:
					inputs[i] = makeFullOnesExt()
				default:
					inputs[i] = makeConstExt()
				}
			}

			_ = board.EvaluateExt(inputs, pool)

			if len(pool.Logs) == 0 {
				t.Fatalf("the pool was not used")
			}

			assert.NoError(t, pool.Errors())
		})
	}

}
