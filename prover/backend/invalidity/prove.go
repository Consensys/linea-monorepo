package invalidity

import (
	"fmt"
	"path"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/linea-monorepo/prover/backend/ethereum"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/circuits/dummy"
	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak"
	keccakDummy "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/protocol/compiler/dummy"
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

// Prove generates a proof for the invalidity circuit.
//
// Three prover modes are supported:
//
//   - Dev: skips all circuit work; directly generates a dummy proof from the
//     functional public inputs. No real constraints are checked.
//
//   - Partial: runs the real zkEVM arithmetization with a check-only
//     compilation suite, then runs the gnark circuit constraint check to
//     verify that the witness satisfies all constraints. A dummy proof is
//     produced. All inputs (traces, Merkle proofs) must be provided.
//
//   - Full (prod): compiles the gnark circuit, loads the trusted setup from
//     disk, and generates a real PLONK proof. All inputs must be provided.
func Prove(cfg *config.Config, req *Request) (*Response, error) {
	profiling.SetMonitorParams(cfg)
	exit.SetIssueHandlingMode(exit.ExitAlways)

	logrus.Infof("Starting invalidity proof for type: %s", req.InvalidityType)

	var (
		c               = invalidity.CircuitInvalidity{}
		setup           circuits.Setup
		serializedProof string
		err             error
		mockCircuitID   circuits.MockCircuitID
		circuitID       circuits.CircuitID
	)

	if err := req.Validate(cfg.Invalidity.ProverMode); err != nil {
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

	funcInput := req.FuncInput()

	if cfg.Invalidity.ProverMode == config.ProverModeDev {
		logrus.Info("Running in DEV mode (generating mock proofs)")
	} else {
		// PARTIAL or PROD MODE - prepare inputs
		isPartial := cfg.Invalidity.ProverMode == config.ProverModePartial
		if isPartial {
			logrus.Infof("Running circuit constraint checking in PARTIAL mode for %s", req.InvalidityType)
		} else {
			logrus.Infof("Running invalidity prover in PROD mode with circuit: %s", circuitID)
		}

		fromAddress := ethereum.GetFrom(tx)

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
			if isPartial {
				// is partial mode compile with dummy suite.
				assigningInputs.KeccakCompiledIOP, assigningInputs.KeccakProof = invalidity.MakeKeccakProofs(tx, cfg.Invalidity.MaxRlpByteSize, keccakDummy.Compile)
			} else {
				// in full mode compile with proper wizard compilation suite.
				assigningInputs.KeccakCompiledIOP, assigningInputs.KeccakProof = invalidity.MakeKeccakProofs(tx, cfg.Invalidity.MaxRlpByteSize, keccak.WizardCompilationParameters()...)
			}
			accountTrieInputs, fromAddressAccount, err := req.AccountTrieInputs()
			if err != nil {
				return nil, fmt.Errorf("could not extract account trie inputs: %w", err)
			}
			// sanity checks
			if fromAddressAccount != fromAddress {
				utils.Panic("from address mismatch: %v != %v", fromAddressAccount, fromAddress)
			}

			if accountTrieInputs.Root != req.ZkParentStateRootHash {
				utils.Panic("account trie root is different from the parent state root: %v != %v", accountTrieInputs.Root, req.ZkParentStateRootHash)
			}
			assigningInputs.AccountTrieInputs = accountTrieInputs

		case invalidity.BadPrecompile, invalidity.TooManyLogs:
			txHash := ethereum.GetTxHash(tx)
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
				BlockHashList:          []linTypes.FullBytes32{},
				ExecDataSchwarzZipfelX: fext.Element{},
				ExecData:               []byte{},
			}

			var zkEvm *zkevm.ZkEvm
			if isPartial {
				zkEvm = zkevm.FullZkEVMInvalidityCheckOnly(&cfg.TracesLimits, cfg)
			} else {
				zkEvm = zkevm.FullZkEvmInvalidity(&cfg.TracesLimits, cfg)
			}
			proof := zkEvm.ProveInner(zkevmWitness)
			if err := zkEvm.VerifyInner(proof); err != nil {
				utils.Panic("zkEVM proof verification failed: %v", err)
			}
			assigningInputs.Zkevm = zkEvm
			assigningInputs.ZkevmWizardProof = proof

		case invalidity.FilteredAddressFrom, invalidity.FilteredAddressTo:
			if isPartial {
				assigningInputs.KeccakCompiledIOP, assigningInputs.KeccakProof = invalidity.MakeKeccakProofs(tx, cfg.Invalidity.MaxRlpByteSize, keccakDummy.Compile)
			} else {
				assigningInputs.KeccakCompiledIOP, assigningInputs.KeccakProof = invalidity.MakeKeccakProofs(tx, cfg.Invalidity.MaxRlpByteSize, keccak.WizardCompilationParameters()...)
			}
			assigningInputs.StateRootHash = req.ZkParentStateRootHash
		}

		if isPartial {
			if err := c.CheckOnly(assigningInputs); err != nil {
				utils.Panic("circuit constraint check failed: %v", err)
			}
			logrus.Info("Circuit constraint check passed")
		} else {
			logrus.Info("Loading setup...")
			if setup, err = circuits.LoadSetup(cfg, circuitID); err != nil {
				return nil, fmt.Errorf("could not load the setup: %w", err)
			}
			serializedProof = c.MakeProof(setup, assigningInputs)
		}
	}

	// Dummy proofs for dev and partial modes
	if cfg.Invalidity.ProverMode == config.ProverModeDev || cfg.Invalidity.ProverMode == config.ProverModePartial {
		srsProvider, err := circuits.NewSRSStore(cfg.PathForSRS())
		if err != nil {
			utils.Panic("error creating SRS store: %v", err)
		}
		setup, err = dummy.MakeUnsafeSetup(srsProvider, mockCircuitID, ecc.BLS12_377.ScalarField())
		if err != nil {
			utils.Panic("error creating unsafe setup: %v", err)
		}
		// // set X field of the dummy circuit (X5 = X^5 + cirID) to the public input. This is used to check the public input against the contract.
		serializedProof = dummy.MakeProof(&setup, funcInput.SumAsField(), mockCircuitID)
	}

	rsp := &Response{
		Transaction:                      tx,
		Signer:                           funcInput.FromAddress,
		TxHash:                           utils.HexEncodeToString(funcInput.TxHash[:]),
		RLPEncodedTx:                     req.RlpEncodedTx,
		ForcedTransactionNumber:          req.ForcedTransactionNumber,
		PrevFtxRollingHash:               req.PrevFtxRollingHash,
		DeadlineBlockHeight:              req.DeadlineBlockHeight,
		InvalidityType:                   req.InvalidityType,
		ZkParentStateRootHash:            req.ZkParentStateRootHash,
		SimulatedExecutionBlockNumber:    req.SimulatedExecutionBlockNumber,
		SimulatedExecutionBlockTimestamp: req.SimulatedExecutionBlockTimestamp,
		ProverVersion:                    cfg.Version,
		Proof:                            serializedProof,
		VerifyingKeyShaSum:               setup.VerifyingKeyDigest(),
		PublicInput:                      linTypes.Bls12377Fr(funcInput.Sum(nil)),
		FtxRollingHash:                   funcInput.FtxRollingHash,
		ProverMode:                       cfg.Invalidity.ProverMode,
	}
	return rsp, nil
}
