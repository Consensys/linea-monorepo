package invalidity

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/dummy"
	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// Prove generates a proof for the invalidity circuit
func Prove(cfg *config.Config, req *Request) (*Response, error) {
	var (
		c               = invalidity.CircuitInvalidity{}
		setup           circuits.Setup
		serializedProof string
		err             error
	)

	if cfg.Invalidity.ProverMode == config.ProverModeDev {

		srsProvider, err := circuits.NewSRSStore(cfg.PathForSRS())
		if err != nil {
			utils.Panic("error creating SRS store: %v", err)
		}

		setup, err = dummy.MakeUnsafeSetup(srsProvider, circuits.MockCircuitIDExecution, ecc.BLS12_377.ScalarField())
		if err != nil {
			utils.Panic("error creating unsafe setup: %v", err)
		}

		serializedProof = dummy.MakeProof(&setup, req.FuncInput().SumAsField(), circuits.MockCircuitIDExecution)

	} else {

		if setup, err = circuits.LoadSetup(cfg, circuits.InvalidityCircuitID); err != nil {
			return nil, fmt.Errorf("could not load the setup: %w", err)
		}
		var (
			kcomp  *wizard.CompiledIOP
			kproof wizard.Proof
		)
		// if circuit is of type BadNonce/BadBalance, we need to create a keccak module
		if req.InvalidityTypes == invalidity.BadNonce || req.InvalidityTypes == invalidity.BadBalance {
			kcomp, kproof = invalidity.MakeKeccakProofs(req.ForcedTransactionPayLoad, MaxRlpByteSize, dummy.Compile)
		}

		serializedProof = c.MakeProof(setup,
			invalidity.AssigningInputs{
				AccountTrieInputs: req.AccountTrie,
				Transaction:       req.ForcedTransactionPayLoad,
				InvalidityType:    req.InvalidityTypes,
			},
			req.FuncInput(), kcomp, kproof)
	}

	rsp := &Response{
		Request:            *req,
		ProverVersion:      cfg.Version,
		Proof:              serializedProof,
		VerifyingKeyShaSum: setup.VerifyingKeyDigest(),
	}
	return rsp, nil
}
