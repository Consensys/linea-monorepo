package plonkinwizard_aligned

import (
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// plonkInWizardAlignedCompilationContext is an internal
type plonkInWizardAlignedCompilationContext struct {
	Query    query.PlonkInWizardAligned
	Data     ifaces.Column
	IsActive ifaces.Column
	IsData   ifaces.Column
}

func compileQuery(comp *wizard.CompiledIOP, q query.PlonkInWizardAligned) {

	var (
		nbPublic      = q.GetNbPublicInputs()
		nbInputPadded = utils.NextPowerOfTwo(nbPublic)
	)

}
