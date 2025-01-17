package conglomeration

import (
	"strings"

	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

const (
	subProofInStatePrefixStr = ".subProof"
)

// PreVortexProverStep is a step replicating the prover of the tmpl at round
// `Round` before the Vortex compilation step. It works by adding columns from
// a wizard proof stored in the prover runtime. The proof is fetched from the
// runtime state. That means the prover step should be run after the proof has
// been attached to the runtime.
type PreVortexProverStep struct {
	Ctxs  []*recursionCtx
	Round int
}

// AssignVortexUAlpha assigns the UAlpha column for all the subproofs. As
// for [PreVortexVerifierStep], this step should be run after the corresponding
// proofs have been added to the runtime states.
type AssignVortexUAlpha struct {
	Ctxs  []*recursionCtx
	Round int
}

// AssignVortexOpenedCols assigns the OpenedCols for all the subproofs. As
// for [PreVortexVerifierStep], this step should be run after the corresponding
// proofs have been added to the runtime states.
type AssignVortexOpenedCols struct {
	Ctxs  []*recursionCtx
	Round int
}

func (pa PreVortexProverStep) Run(run *wizard.ProverRuntime) {
	for _, ctx := range pa.Ctxs {

		var (
			prefix        = ctx.Translator.Prefix
			proof         = run.State.MustGet(prefix + subProofInStatePrefixStr).(wizard.Proof)
			queriesParams = ctx.QueryParams[pa.Round]
			colums        = ctx.Columns[pa.Round]
		)

		for _, col := range colums {
			name := unprefix(prefix, col.GetColID())
			run.AssignColumn(col.GetColID(), proof.Messages.MustGet(name))
		}

		for _, param := range queriesParams {
			name := unprefix(prefix, param.Name())
			run.QueriesParams.InsertNew(param.Name(), proof.QueriesParams.MustGet(name))
		}
	}
}

func (pa AssignVortexUAlpha) Run(run *wizard.ProverRuntime) {
	for _, ctx := range pa.Ctxs {
		// Since all the context of the pcs is translated, this does not
		// need to run over a translated prover runtime.
		ctx.PcsCtx.ComputeLinearComb(run)
	}
}

func (pa AssignVortexOpenedCols) Run(run *wizard.ProverRuntime) {
	for _, ctx := range pa.Ctxs {
		// Since all the context of the pcs is translated, this does not
		// need to run over a translated prover runtime.
		ctx.PcsCtx.OpenSelectedColumns(run)
	}
}

func unprefix[T ~string](prefix string, name T) T {
	p, n := string(prefix), string(name)
	r := strings.TrimPrefix(n, p)
	return T(r)
}
