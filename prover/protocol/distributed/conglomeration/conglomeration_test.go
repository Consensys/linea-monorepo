package conglomeration

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/dummy"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

func TestConglomerationPureVortex(t *testing.T) {

	var (
		numCol         = 16
		numRow         = 16
		numProof       = 16
		a              []ifaces.Column
		u              query.UnivariateEval
		sisParams      = &ringsis.Params{LogTwoBound: 16, LogTwoDegree: 1}
		vortexCompFunc = vortex.Compile(2, vortex.WithSISParams(sisParams), vortex.ForceNumOpenedColumns(2))
	)

	define := func(builder *wizard.Builder) {
		for i := 0; i < numCol; i++ {
			a = append(a, builder.RegisterCommit(ifaces.ColIDf("a-%v", i), numRow))
		}
		u = builder.CompiledIOP.InsertUnivariate(0, "u", a)
	}

	prover := func(k int) func(run *wizard.ProverRuntime) {
		return func(run *wizard.ProverRuntime) {
			ys := make([]field.Element, len(a))
			for i := range a {
				y := field.NewElement(uint64(i + k))
				run.AssignColumn(a[i].GetColID(), smartvectors.NewConstant(y, numCol))
				ys = append(ys, y)
			}
			run.AssignUnivariate(u.QueryID, field.NewElement(0), ys...)
		}
	}

	var (
		tmpl                 = wizard.Compile(define, vortexCompFunc)
		congDef, ctxsPHolder = ConglomerateDefineFunc(tmpl, numProof)
		cong                 = wizard.Compile(congDef, dummy.CompileAtProverLvl)
		ctxs                 = *ctxsPHolder
		lastRound            = ctxs[0].LastRound
	)

	witnesses := make([]Witness, numProof)
	for i := range witnesses {
		runtime := wizard.RunProverUntilRound(tmpl, prover(i), lastRound+1)
		witnesses[i] = ExtractWitness(runtime)
	}

	proof := wizard.Prove(cong, ProveConglomeration(ctxs, witnesses))
	err := wizard.Verify(cong, proof)

	require.NoError(t, err)

}
