package publicInput

import (
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/zkevm/arithmetization"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/statesummary"
	"github.com/ethereum/go-ethereum/common"
)

// Settings for PublicInput module (stub)
type Settings struct {
	Name string
}

// PublicInput module (stub)
type PublicInput struct{}

// NewPublicInputZkEVM creates a new PublicInput module (stub)
func NewPublicInputZkEVM(comp *wizard.CompiledIOP, settings *Settings, ss *statesummary.Module, a *arithmetization.Arithmetization) PublicInput {
	return PublicInput{}
}

// Assign assigns the public input (stub)
func (pub *PublicInput) Assign(run *wizard.ProverRuntime, l2BridgeAddress common.Address, blockHashList []types.FullBytes32) {
	// stub - do nothing
}

