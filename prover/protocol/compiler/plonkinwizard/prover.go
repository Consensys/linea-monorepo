package plonkinwizard

import (
	plonk "github.com/consensys/linea-monorepo/prover/protocol/compiler/plonkinwizard/internal"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

type proverAction struct {
	Ctx *plonk.CompilationCtx
	Q   *query.PlonkInWizard
}

func (a *proverAction) Run(run *wizard.ProverRuntime) error {
	a.Ctx.GetPlonkProverAction().Run(a.Ctx, a.Q)
}

func 
