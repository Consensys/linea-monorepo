package symbolic_test

import (
	"fmt"
	"testing"

	sv "github.com/consensys/accelerated-crypto-monorepo/maths/common/smartvectors"
	"github.com/consensys/accelerated-crypto-monorepo/maths/common/vector"
	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/arithmetics"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/logdata"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/specialqueries"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/query"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/variables"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/zkevm"
	"github.com/consensys/accelerated-crypto-monorepo/zkevm/define"
)

/*
Returns a compiled IOP
*/
func compiledIOP() *wizard.CompiledIOP {
	return wizard.Compile(
		zkevm.WrapDefine(define.ZkEVMDefine, nil),
		logdata.Log("defined"),
		specialqueries.RangeProof,
		logdata.Log("ranges"),
		specialqueries.LogDerivativeLookupCompiler,
		specialqueries.CompilePermutations,
		logdata.Log("inclusions-permutations"),
		arithmetics.CompileLocal,
		logdata.Log("locals"),
	)
}

/*
Benchmark the expressions of the global constraint of the zk-EVM
*/
func BenchmarkMergedExpression(b *testing.B) {

	compiled := compiledIOP()
	size := 1 << 18

	for _, qName := range compiled.QueriesNoParams.AllKeys() {
		if compiled.QueriesNoParams.IsIgnored(qName) {
			// Ignore compiled queries
			continue
		}

		q := compiled.QueriesNoParams.Data(qName).(query.GlobalConstraint)
		boarded := q.Board()
		metadatas := boarded.ListVariableMetadata()
		inputs := make([]sv.SmartVector, len(metadatas))
		randVec := sv.NewRegular(vector.Rand(size))
		constVec := sv.NewConstant(field.NewElement(42), size)

		for i, m := range metadatas {
			switch m.(type) {
			case ifaces.Column:
				inputs[i] = randVec
			case coin.Info, *ifaces.Accessor:
				inputs[i] = constVec
			case variables.X, variables.PeriodicSample:
				inputs[i] = randVec
			}
		}

		b.Run(fmt.Sprintf("evaluate expression of %v for size %v", qName, size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				boarded.Evaluate(inputs)
			}
		})

	}

}
