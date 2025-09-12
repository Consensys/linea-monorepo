package invalidity

import (
	"bytes"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/linea-monorepo/prover/backend/ethereum"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/dummy"
	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Prove generates a proof for the invalidity circuit
func Prove(cfg *config.Config, req *Request) (*Response, error) {
	var (
		c               = invalidity.CircuitInvalidity{}
		setup           circuits.Setup
		serializedProof string
		err             error
		txData          types.TxData
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

		if txData, err = ethereum.DecodeTxFromBytes(bytes.NewReader(req.RlpEncodedTx)); err != nil {
			return nil, fmt.Errorf("could not decode the RlpEncodedTx %w", err)
		}

		serializedProof = c.MakeProof(setup,
			invalidity.AssigningInputs{
				RlpEncodedTx:      req.RlpEncodedTx,
				Transaction:       types.NewTx(txData),
				AccountTrieInputs: req.AccountTrie,
				FromAddress:       common.Address(req.FromAddresses),
				InvalidityType:    req.InvalidityTypes,
				FuncInputs:        *req.FuncInput(),
				MaxRlpByteSize:    cfg.Invalidity.MaxRlpByteSize,
			},
		)
	}

	rsp := &Response{
		Request:            *req,
		ProverVersion:      cfg.Version,
		Proof:              serializedProof,
		VerifyingKeyShaSum: setup.VerifyingKeyDigest(),
	}
	return rsp, nil
}
