package invalidity

import (
	"fmt"
	"path"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/linea-monorepo/prover/backend/ethereum"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/dummy"
	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	wizardk "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/exit"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	linTypes "github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// Prove generates a proof for the invalidity circuit
func Prove(cfg *config.Config, req *Request, compilationSuite ...func(*wizardk.CompiledIOP)) (*Response, error) {
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
	case invalidity.FilteredAddressFrom, invalidity.FilteredAddressTo:
		mockCircuitID = circuits.MockCircuitIDInvalidityFilteredAddress
		circuitID = circuits.InvalidityFilteredAddressCircuitID
	default:
		return nil, fmt.Errorf("unsupported invalidity type: %s", req.InvalidityType)
	}

	tx, err := ethereum.RlpDecodeWithSignature(req.RlpEncodedTx)
	if err != nil {
		return nil, fmt.Errorf("could not decode the RlpEncodedTx: %w", err)
	}

	// Compute functional inputs (includes TxHash and FtxRollingHash)
	funcInput := req.FuncInput()

	if cfg.Invalidity.ProverMode == config.ProverModeDev || cfg.Invalidity.ProverMode == config.ProverModePartial {
		// DEV/PARTIAL MODE - uses dummy/mock proofs
		if cfg.Invalidity.ProverMode == config.ProverModePartial {

			// For BadPrecompile/TooManyLogs, run constraint checking on zkEVM
			if req.InvalidityType == invalidity.BadPrecompile || req.InvalidityType == invalidity.TooManyLogs {
				logrus.Info("Running zkEVM constraint checking for BadPrecompile/TooManyLogs in the PARTIAL mode")
				txHash := ethereum.GetTxHash(tx)
				txSignature := ethereum.GetJsonSignature(tx)

				zkevmWitness := &zkevm.Witness{
					ExecTracesFPath: path.Join(cfg.Execution.ConflatedTracesDir, req.ConflatedExecutionTracesFile),
					SMTraces:        req.ZkStateMerkleProof,
					TxSignatures:    []ethereum.Signature{txSignature},
					TxHashes:        [][32]byte{txHash},
					L2BridgeAddress: cfg.Layer2.MsgSvcContract,
					ChainID:         cfg.Layer2.ChainID,
					BlockHashList:   []linTypes.FullBytes32{},
				}

				traces := &cfg.TracesLimits
				partial := zkevm.FullZkEVMCheckOnly(traces, cfg)
				proof := partial.ProveInner(zkevmWitness)
				if err := partial.VerifyInner(proof); err != nil {
					utils.Panic("zkEVM constraint check failed: %v", err)
				}
				logrus.Info("zkEVM constraint check passed")
			}
		} else {
			logrus.Info("has fallen into the DEV mode (generating mock proofs)")
		}

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

		// Prepare AssigningInputs (common fields for all invalidity types)
		assigningInputs := invalidity.AssigningInputs{
			RlpEncodedTx:   ethereum.EncodeTxForSigning(tx),
			Transaction:    tx,
			FromAddress:    common.Address(fromAddress),
			InvalidityType: req.InvalidityType,
			FuncInputs:     *funcInput,
			MaxRlpByteSize: cfg.Invalidity.MaxRlpByteSize,
		}

		switch req.InvalidityType {
		case invalidity.BadNonce, invalidity.BadBalance:
			// BadNonce/BadBalance only need AccountTrieInputs and Keccak proof
			assigningInputs.KeccakCompiledIOP, assigningInputs.KeccakProof = invalidity.MakeKeccakProofs(assigningInputs.Transaction, assigningInputs.MaxRlpByteSize, compilationSuite...)
			logrus.Info("Extracting account trie inputs for BadNonce/BadBalance...")
			accountTrieInputs, fromAddressAccount, err = req.AccountTrieInputs()
			if err != nil {
				return nil, fmt.Errorf("could not extract account trie inputs: %w", err)
			}

			// sanity check on the account trie inputs
			if fromAddressAccount != fromAddress {
				utils.Panic("from address mismatch: %v != %v", fromAddressAccount, fromAddress)
			}
			assigningInputs.AccountTrieInputs = accountTrieInputs

		case invalidity.BadPrecompile, invalidity.TooManyLogs:
			// BadPrecompile/TooManyLogs need full zkEVM proof
			logrus.Info("Building zkEVM witness for BadPrecompile/TooManyLogs...")
			txHash := ethereum.GetTxHash(tx) // unsigned tx hash
			txSignature := ethereum.GetJsonSignature(tx)

			zkevmWitness := &zkevm.Witness{
				ExecTracesFPath:        path.Join(cfg.Execution.ConflatedTracesDir, req.ConflatedExecutionTracesFile),
				SMTraces:               req.ZkStateMerkleProof,
				TxSignatures:           []ethereum.Signature{txSignature},
				TxHashes:               [][32]byte{txHash},
				L2BridgeAddress:        cfg.Layer2.MsgSvcContract,
				ChainID:                cfg.Layer2.ChainID,
				BaseFee:                cfg.Layer2.BaseFee,
				CoinBase:               linTypes.EthAddress(cfg.Layer2.CoinBase),
				BlockHashList:          []linTypes.FullBytes32{}, // Not used for invalidity proofs
				ExecDataSchwarzZipfelX: fext.Element{},           // not used for invalidity proofs
				ExecData:               []byte{},                 // not used for invalidity proofs
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

			assigningInputs.Zkevm = fullZkEvm
			assigningInputs.ZkevmWizardProof = proof

		case invalidity.FilteredAddressFrom, invalidity.FilteredAddressTo:
			logrus.Info("Processing FilteredAddress invalidity type...")
			assigningInputs.KeccakCompiledIOP, assigningInputs.KeccakProof = invalidity.MakeKeccakProofs(assigningInputs.Transaction, assigningInputs.MaxRlpByteSize, compilationSuite...)
			assigningInputs.StateRootHash = req.ZkParentStateRootHash
		}

		serializedProof = c.MakeProof(setup, assigningInputs)
	}

	rsp := &Response{
		Transaction:        tx,
		Signer:             funcInput.FromAddress,
		TxHash:             utils.HexEncodeToString(funcInput.TxHash[:]),
		Request:            *req,
		ProverVersion:      cfg.Version,
		Proof:              serializedProof,
		VerifyingKeyShaSum: setup.VerifyingKeyDigest(),
		PublicInput:        linTypes.Bls12377Fr(funcInput.Sum(nil)),
		FtxRollingHash:     funcInput.FtxRollingHash,
	}
	return rsp, nil
}
