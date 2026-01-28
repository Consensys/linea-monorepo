package invalidity

import (
	"fmt"
	"path"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/linea-monorepo/prover/backend/ethereum"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/dummy"
	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/utils"
	linTypes "github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Prove generates a proof for the invalidity circuit
func Prove(cfg *config.Config, req *Request) (*Response, error) {
	var (
		c                  = invalidity.CircuitInvalidity{}
		setup              circuits.Setup
		serializedProof    string
		err                error
		mockCircuitID      circuits.MockCircuitID
		circuitID          circuits.CircuitID
		accountTrieInputs  invalidity.AccountTrieInputs
		fromAddressAccount linTypes.EthAddress
	)

	switch req.InvalidityType {
	case invalidity.BadNonce, invalidity.BadBalance:
		mockCircuitID = circuits.MockCircuitIDInvalidityNonceBalance
		circuitID = circuits.InvalidityNonceBalanceCircuitID
	case invalidity.BadPrecompile, invalidity.TooManyLogs:
		mockCircuitID = circuits.MockCircuitIDInvalidityPrecompileLogs
		circuitID = circuits.InvalidityPrecompileLogsCircuitID
	default:
		return nil, fmt.Errorf("unsupported invalidity type: %s", req.InvalidityType)
	}

	// Decode the signed transaction
	tx := new(types.Transaction)
	if err = tx.UnmarshalBinary(req.RlpEncodedTx); err != nil {
		return nil, fmt.Errorf("could not decode the RlpEncodedTx %w", err)
	}

	// Compute functional inputs (includes TxHash and FtxRollingHash)
	funcInput := req.FuncInput()

	if cfg.Invalidity.ProverMode == config.ProverModeDev {
		// DEV MODE - uses dummy/mock proofs

		srsProvider, err := circuits.NewSRSStore(cfg.PathForSRS())
		if err != nil {
			utils.Panic("error creating SRS store: %v", err)
		}

		setup, err = dummy.MakeUnsafeSetup(srsProvider, mockCircuitID, ecc.BLS12_377.ScalarField())
		if err != nil {
			utils.Panic("error creating unsafe setup: %v", err)
		}

		serializedProof = dummy.MakeProof(&setup, funcInput.SumAsField(), mockCircuitID)

	} else {
		// PROD MODE - uses real proofs

		if setup, err = circuits.LoadSetup(cfg, circuitID); err != nil {
			return nil, fmt.Errorf("could not load the setup: %w", err)
		}

		accountTrieInputs, fromAddressAccount, err = req.AccountTrieInputs()
		if err != nil {
			utils.Panic("could not extract account trie inputs: %v", err)
		}

		// sanity check on the account trie inputs
		fromAddress := ethereum.GetFrom(tx)
		if fromAddressAccount != fromAddress {
			utils.Panic("from address mismatch: %v != %v", fromAddressAccount, fromAddress)
		}

		// Build zkevm.Witness with the single invalid transaction
		txHash := ethereum.GetTxHash(tx)
		txSignature := ethereum.GetJsonSignature(tx)

		zkevmWitness := &zkevm.Witness{
			ExecTracesFPath: path.Join(cfg.Execution.ConflatedTracesDir, req.ConflatedExecutionTracesFile),
			SMTraces:        req.ZkStateMerkleProof,
			TxSignatures:    []ethereum.Signature{txSignature},
			TxHashes:        [][32]byte{txHash},
			L2BridgeAddress: cfg.Layer2.MsgSvcContract,
			ChainID:         cfg.Layer2.ChainID,
			// BlockHashList:   nil, // Not used for invalidity proofs
		}

		traces := &cfg.TracesLimits
		fullZkEvm := zkevm.FullZkEvm(traces, cfg)
		// Generates the inner-proof and sanity-check it so that we ensure that
		// the prover nevers outputs invalid proofs.
		proof := fullZkEvm.ProveInner(zkevmWitness)

		serializedProof = c.MakeProof(setup,
			invalidity.AssigningInputs{
				RlpEncodedTx:      req.RlpEncodedTx,
				Transaction:       tx,
				AccountTrieInputs: accountTrieInputs,
				FromAddress:       common.Address(fromAddress),
				InvalidityType:    req.InvalidityType,
				FuncInputs:        *funcInput,
				MaxRlpByteSize:    cfg.Invalidity.MaxRlpByteSize,
				Zkevm:             fullZkEvm,
				ZkevmWizardProof:  proof,
			})
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
