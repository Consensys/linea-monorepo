package invalidity

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/config"
	public_input "github.com/consensys/linea-monorepo/prover/public-input"
)

func Prove(cfg *config.Config, req *Request) (*Response, error) {
	var (
		c     = invalidity.CircuitInvalidity{}
		setup circuits.Setup
		err   error
	)

	// out := CraftProverOutput(cfg, req)

	if setup, err = circuits.LoadSetup(cfg, circuits.InvalidityCircuitID); err != nil {
		return nil, fmt.Errorf("could not load the setup: %w", err)
	}
	//@azam add the details for different ProverModes
	serializedProof := c.MakeProof(setup,
		invalidity.AssigningInputs{
			AccountTrieInputs: req.AccountTri,
			Transaction:       req.ForcedTransactionPayLoad,
			InvalidityType:    req.InvalidityTypes,
		},
		public_input.Invalidity{
			TxNumber:            req.ForcedTransactionNumbers,
			FromAddress:         req.FromAddresses,
			ExpectedBlockHeight: req.ExpectedBlockHeights,
			StateRootHash:       req.StateRootHash,
		})

	resp := &Response{
		Request:            *req,
		ProverVersion:      cfg.Version,
		Proof:              serializedProof,
		VerifyingKeyShaSum: setup.VerifyingKeyDigest(),
	}
	return resp, nil
}
