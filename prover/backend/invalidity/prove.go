package invalidity

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/dummy"
	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/utils"
	linTypes "github.com/consensys/linea-monorepo/prover/utils/types"
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
	)

	// Decode the signed transaction
	tx := new(types.Transaction)
	if err = tx.UnmarshalBinary(req.RlpEncodedTx); err != nil {
		return nil, fmt.Errorf("could not decode the RlpEncodedTx %w", err)
	}

	// Compute functional inputs (includes TxHash and FtxRollingHash)
	funcInput := req.FuncInput()

	if cfg.Invalidity.ProverMode == config.ProverModeDev {

		srsProvider, err := circuits.NewSRSStore(cfg.PathForSRS())
		if err != nil {
			utils.Panic("error creating SRS store: %v", err)
		}

		setup, err = dummy.MakeUnsafeSetup(srsProvider, circuits.MockCircuitIDExecution, ecc.BLS12_377.ScalarField())
		if err != nil {
			utils.Panic("error creating unsafe setup: %v", err)
		}

		serializedProof = dummy.MakeProof(&setup, funcInput.SumAsField(), circuits.MockCircuitIDExecution)

	} else {

		if setup, err = circuits.LoadSetup(cfg, circuits.InvalidityCircuitID); err != nil {
			return nil, fmt.Errorf("could not load the setup: %w", err)
		}

		// Extract AccountTrieInputs from the Shomei traces for the sender
		accountTrieInputs, fromAddress, err := req.AccountTrieInputs()
		if err != nil {
			return nil, fmt.Errorf("could not extract account trie inputs: %w", err)
		}

		serializedProof = c.MakeProof(setup,
			invalidity.AssigningInputs{
				RlpEncodedTx:      req.RlpEncodedTx,
				Transaction:       tx,
				AccountTrieInputs: accountTrieInputs,
				FromAddress:       common.Address(fromAddress),
				InvalidityType:    req.InvalidityTypes,
				FuncInputs:        *funcInput,
				MaxRlpByteSize:    cfg.Invalidity.MaxRlpByteSize,
			},
		)
	}

	rsp := &Response{
		Transaction:        tx,
		Signer:             funcInput.FromAddress,
		TxHash:             utils.HexEncodeToString(funcInput.TxHash[:]),
		Request:            *req,
		ProverVersion:      cfg.Version,
		Proof:              serializedProof,
		VerifyingKeyShaSum: setup.VerifyingKeyDigest(),
		PublicInput:        linTypes.Bytes32(funcInput.Sum(nil)),
		FtxRollingHash:     funcInput.FtxRollingHash,
	}
	return rsp, nil
}
