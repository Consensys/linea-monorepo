package publicInput

import (
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/ecdsa"
	invalidity "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/invalidity_pi"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/statemanager/statesummary"
	"github.com/ethereum/go-ethereum/common"
)

// PublicInputWrapper wraps either a PublicInput (for validity proofs) or
// an InvalidityPI (for invalidity proofs) depending on the proof mode.
type PublicInputs struct {
	ExecutionPI    *PublicInput
	InvalidityPI   *invalidity.InvalidityPI
	IsInvalidityPI bool
}

// NewPublicInput creates a new PublicInputWrapper.
// If isInvalidityPI is true, it creates an InvalidityPI module.
// Otherwise, it creates a standard PublicInput module.
func NewPublicInput(comp *wizard.CompiledIOP, isInvalidityPI bool, settings *Settings, ss *statesummary.Module, ecdsaModule *ecdsa.EcdsaZkEvm) *PublicInputs {
	if isInvalidityPI {
		return &PublicInputs{
			ExecutionPI:    nil,
			InvalidityPI:   invalidity.NewInvalidityPI(comp, ecdsaModule, ss),
			IsInvalidityPI: true,
		}
	}
	pi := NewPublicInputZkEVM(comp, settings, ss)
	return &PublicInputs{
		ExecutionPI:    &pi,
		InvalidityPI:   nil,
		IsInvalidityPI: false,
	}
}

// Assign assigns values to the public input columns.
// For execution proofs, blockHashList is required.
// For invalidity proofs, blockHashList is ignored.
func (p *PublicInputs) Assign(run *wizard.ProverRuntime, l2BridgeAddress common.Address, blockHashList []types.FullBytes32) {
	if p.IsInvalidityPI {
		p.InvalidityPI.Assign(run, l2BridgeAddress)
	} else {
		p.ExecutionPI.Assign(run, l2BridgeAddress, blockHashList)
	}
}
