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
	"github.com/consensys/linea-monorepo/prover/utils/exit"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	linTypes "github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// Prove generates a proof for the invalidity circuit
func Prove(cfg *config.Config, req *Request) (*Response, error) {
	// Set up profiling and exit handling
	profiling.SetMonitorParams(cfg)
	exit.SetIssueHandlingMode(exit.ExitAlways)

	logrus.Infof("Starting invalidity proof for type: %s", req.InvalidityType)

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

	// Validate the request before processing
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	switch req.InvalidityType {
	case invalidity.BadNonce, invalidity.BadBalance:
		mockCircuitID = circuits.MockCircuitIDInvalidityNonceBalance
		circuitID = circuits.InvalidityNonceBalanceCircuitID
	case invalidity.BadPrecompile, invalidity.TooManyLogs:
		mockCircuitID = circuits.MockCircuitIDInvalidityPrecompileLogs
		circuitID = circuits.InvalidityPrecompileLogsCircuitID
	case invalidity.FilteredAddresses:
		// FilteredAddresses case: no merkle proof needed, just verify the address is filtered
		// TODO: Implement FilteredAddresses circuit IDs once available
		return nil, fmt.Errorf("FilteredAddresses invalidity type is not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported invalidity type: %s", req.InvalidityType)
	}

	tx, err := ethereum.RlpDecodeWithSignature(req.RlpEncodedTx)
	if err != nil {
		return nil, fmt.Errorf("could not decode the RlpEncodedTx: %w", err)
	}

	// Compute functional inputs (includes TxHash and FtxRollingHash)
	funcInput := req.FuncInput()

	if cfg.Invalidity.ProverMode == config.ProverModeDev {
		// DEV MODE - uses dummy/mock proofs
		logrus.Info("Running invalidity prover in DEV mode")

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
		logrus.Infof("Running invalidity prover in PROD mode with circuit: %s", circuitID)

		logrus.Info("Loading setup...")
		if setup, err = circuits.LoadSetup(cfg, circuitID); err != nil {
			return nil, fmt.Errorf("could not load the setup: %w", err)
		}

		fromAddress := ethereum.GetFrom(tx)

		// Only extract AccountTrieInputs for BadNonce/BadBalance cases
		if req.InvalidityType == invalidity.BadNonce || req.InvalidityType == invalidity.BadBalance {
			accountTrieInputs, fromAddressAccount, err = req.AccountTrieInputs()
			if err != nil {
				return nil, fmt.Errorf("could not extract account trie inputs: %w", err)
			}

			// sanity check on the account trie inputs
			if fromAddressAccount != fromAddress {
				utils.Panic("from address mismatch: %v != %v", fromAddressAccount, fromAddress)
			}
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
			BlockHashList:   []linTypes.FullBytes32{}, // Not used for invalidity proofs, so we leave it empty
		}

		traces := &cfg.TracesLimits
		logrus.Info("Getting Full ZkEVM for invalidity...")
		fullZkEvm := zkevm.FullZkEvmInvalidity(traces, cfg)

		// Generates the inner-proof and sanity-check it
		logrus.Info("Generating inner proof...")
		proof := fullZkEvm.ProveInner(zkevmWitness)

		// Verify the inner proof to ensure prover never outputs invalid proofs
		logrus.Info("Verifying inner proof...")
		if err := fullZkEvm.VerifyInner(proof); err != nil {
			utils.Panic("inner proof verification failed: %v", err)
		}

		serializedProof = c.MakeProof(setup,
			invalidity.AssigningInputs{
				RlpEncodedTx:      ethereum.EncodeTxForSigning(tx),
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
