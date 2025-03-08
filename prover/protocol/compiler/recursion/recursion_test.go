package recursion

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/globalcs"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/localcs"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logderivativesum"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/univariates"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLookup(t *testing.T) {

	logrus.SetLevel(logrus.FatalLevel)

	define1 := func(bui *wizard.Builder) {

		var (
			a = bui.RegisterCommit("A", 8)
			b = bui.RegisterCommit("B", 8)
		)

		bui.Inclusion("Q", []ifaces.Column{a}, []ifaces.Column{b})
	}

	prove1 := func(run *wizard.ProverRuntime) {
		run.AssignColumn("A", smartvectors.ForTest(1, 2, 3, 4, 5, 6, 7, 8))
		run.AssignColumn("B", smartvectors.ForTest(1, 2, 3, 4, 5, 6, 7, 8))
	}

	suites := [][]func(*wizard.CompiledIOP){
		{
			logderivativesum.CompileLookups,
			localcs.Compile,
			globalcs.Compile,
			univariates.CompileLocalOpening,
			univariates.Naturalize,
			univariates.MultiPointToSinglePoint(8),
			vortex.Compile(2, vortex.ForceNumOpenedColumns(4), vortex.WithSISParams(&ringsis.StdParams), vortex.PremarkAsSelfRecursed()),
		},
	}

	for i, s := range suites {

		t.Run(fmt.Sprintf("case-%v", i), func(t *testing.T) {

			comp1 := wizard.Compile(define1, s...)
			var recCtx *Recursion

			define2 := func(build2 *wizard.Builder) {
				recCtx = DefineRecursionOf("test", build2.CompiledIOP, comp1, true, 1)
			}

			comp2 := wizard.Compile(define2, dummy.CompileAtProverLvl)

			proverRuntime := wizard.RunProverUntilRound(comp1, prove1, 6)
			witness1 := ExtractWitness(proverRuntime)

			prove2 := func(run *wizard.ProverRuntime) {
				recCtx.Assign(run, []Witness{witness1})
			}

			proof2 := wizard.Prove(comp2, prove2)
			assert.NoErrorf(t, wizard.Verify(comp2, proof2), "invalid proof")
		})
	}
}
