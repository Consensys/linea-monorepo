package inclusion_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/distributed"
	"github.com/consensys/linea-monorepo/prover/protocol/distributed/compiler/inclusion"
	modulediscoverer "github.com/consensys/linea-monorepo/prover/protocol/distributed/module_discoverer"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
)

// It tests if the expression is properly distributed among modules.
// It does not test the prover steps.
func TestDistributedLogDerivSumExpr(t *testing.T) {

	var (
		// columns that are used in the cross queries.
		crossCols        [2]ifaces.Column
		define0, define1 func(b *wizard.Builder)
	)

	// moduleComp0
	define0 = func(b *wizard.Builder) {
		moduleCol := b.CompiledIOP.InsertCommit(0, "module0.col0", 4)
		crossCols[0] = b.CompiledIOP.InsertCommit(0, "module0.col1", 4)
		// b.CompiledIOP.InsertInclusion(0, "moduel0.lookup0", []ifaces.Column{crossCols[1]}, []ifaces.Column{moduleCol})
		b.CompiledIOP.InsertGlobal(0, "module0.global0",
			symbolic.Sub(moduleCol, symbolic.Mul(2, crossCols[0])))
	}

	// moduleComp1
	define1 = func(b *wizard.Builder) {
		moduleCol := b.CompiledIOP.InsertCommit(0, "module1.col0", 8)
		crossCols[1] = b.CompiledIOP.InsertCommit(0, "module1.col1", 8)
		moduleCol1 := b.CompiledIOP.InsertCommit(0, "module1.col2", 8)
		b.CompiledIOP.InsertInclusion(0, "module1.lookup0", []ifaces.Column{moduleCol}, []ifaces.Column{crossCols[0]})
		b.CompiledIOP.InsertGlobal(0, "module1.global0",
			symbolic.Sub(moduleCol1, symbolic.Mul(2, moduleCol)))
	}

	// initialComp
	define := func(b *wizard.Builder) {
		define0(b)
		define1(b)
	}
	// in initialComp replace inclusion queries with a global LogDerivativeSum
	initialComp := wizard.Compile(define, distributed.IntoLogDerivativeSum)
	moduleComp0 := wizard.Compile(define0)
	moduleComp1 := wizard.Compile(define1)

	// Initialise the period separating module discoverer
	disc := modulediscoverer.PeriodSeperatingModuleDiscoverer{}
	disc.Analyze(initialComp)

	// distribute the shares to modules.
	inclusion.DistributeLogDerivativeSum(initialComp, moduleComp0, "module0", &disc)
	inclusion.DistributeLogDerivativeSum(initialComp, moduleComp1, "module1", &disc)

}
