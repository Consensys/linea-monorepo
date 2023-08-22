package prover

import (
	"github.com/consensys/accelerated-crypto-monorepo/glue"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/zkevm"
	"github.com/consensys/accelerated-crypto-monorepo/zkevm/statemanager"
)

// Prover options collects parameters for the prover.
// If a function pointer is set to nil, we understand it
// as the option being disabled.
type ProverOptions struct {
	// Enables Keccak proving
	Keccak struct {
		Enabled bool
	}
	// Enables ECDSA verification
	Ecdsa struct {
		Enabled      bool
		SigExtractor glue.TxSignatureExtractor
	}
	// Enables state-manager verification
	StateManager struct {
		AssignFunc func(run *wizard.ProverRuntime)
		Enabled    bool
	}
}

// NewProverOptions creates a new ProverOptions
func NewProverOptions() *ProverOptions {
	return &ProverOptions{}
}

// WithKeccak adds keccak parameters to the ProverOptions
func (po *ProverOptions) WithKeccak() *ProverOptions {
	po.Keccak.Enabled = true
	return po
}

// WithStateManager adds state manager parameters to the ProverOptions
func (po *ProverOptions) WithStateManager(traces [][]any) *ProverOptions {
	po.StateManager.Enabled = true
	po.StateManager.AssignFunc = func(run *wizard.ProverRuntime) {
		statemanager.AssignStateManagerMerkleProof(run, traces)
	}
	return po
}

// WithEcdsa adds ecdsa parameters to the ProverOptions
func (po *ProverOptions) WithEcdsa(sigExtractor glue.TxSignatureExtractor) *ProverOptions {
	po.Ecdsa.Enabled = true
	po.Ecdsa.SigExtractor = sigExtractor
	return po
}

// Returns the optional steps of the prover
func (po *ProverOptions) ApplyProverSteps(
	mainProverStep func(run *wizard.ProverRuntime),
) func(*wizard.ProverRuntime) {

	return func(run *wizard.ProverRuntime) {
		// First run the initial prover step
		mainProverStep(run)

		// Keccak does not seem to need an specific
		// assignment function to run

		// Optional run the ecdsa assignment
		if po.Ecdsa.Enabled {
			panic("not implemented yet")
			// po.Ecdsa.AssignFunc(run)
		}

		// Optional run the state-management assignment
		if po.StateManager.Enabled {
			po.StateManager.AssignFunc(run)
		}
	}
}

// Returns the zkevm.Options for the
func (po *ProverOptions) AppendToZkEvmOptions(ops ...zkevm.Option) []zkevm.Option {

	// Optional run the keccak assignment
	if po.Keccak.Enabled {
		ops = append(ops, zkevm.WithKeccak())
	}

	// Optional run the state-management assignment
	if po.StateManager.Enabled {
		ops = append(ops, zkevm.WithStateManager)
	}

	// Optional run the ecdsa assignment
	if po.Ecdsa.Enabled {
		ops = append(ops, zkevm.WithECDSA(po.Ecdsa.SigExtractor))
	}

	return ops
}
