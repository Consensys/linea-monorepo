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
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/exit"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	linTypes "github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/consensys/linea-monorepo/prover/zkevm"
	invalidityPI "github.com/consensys/linea-monorepo/prover/zkevm/prover/publicInput/invalidity_pi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

// Prove generates a proof for the invalidity circuit.
//
// Three prover modes are supported:
//
//   - Dev: skips all circuit work; directly generates a dummy proof from the
//     functional public inputs. No real constraints are checked.
//
//   - Partial: runs the gnark circuit constraint check (CheckOnly) to verify
//     that the witness satisfies all constraints, then produces a dummy proof.
//     For BadNonce/BadBalance, real keccak proofs and Merkle proofs are used.
//     For FilteredAddress, real keccak proofs are used.
//     For BadPrecompile/TooManyLogs, it goes to a test mode where a mock wizard is built via
//     MockZkevmArithCols (no .lt file or Shomei traces needed).
//
//   - Full (prod): compiles the gnark circuit, loads the trusted setup from
//     disk, and generates a real PLONK proof. All inputs must be provided,
//     including conflated execution traces and Shomei traces for
//     BadPrecompile/TooManyLogs.
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
				assigningInputs.KeccakCompiledIOP, assigningInputs.KeccakProof = invalidity.MakeKeccakProofs(tx, cfg.Invalidity.MaxRlpByteSize, keccakDummy.Compile)
			} else {
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
			if isPartial {
				// Partial mode: use MockZkevmArithCols (no real traces needed)
				logrus.Info("Partial mode: trace files are ignored, using MockZkevmArithCols for BadPrecompile/TooManyLogs")
				mockInputs := buildMockZkevmInputs(tx, fromAddress, req)
				comp, proof := invalidityPI.MockZkevmArithCols(mockInputs)
				assigningInputs.Zkevm = &zkevm.ZkEvm{InitialCompiledIOP: comp}
				assigningInputs.ZkevmWizardProof = proof
			} else {
				// Full/prod mode: traces already validated by Validate()
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

				fullZkEvm := zkevm.FullZkEvmInvalidity(&cfg.TracesLimits, cfg)
				proof := fullZkEvm.ProveInner(zkevmWitness)
				if err := fullZkEvm.VerifyInner(proof); err != nil {
					utils.Panic("zkEVM proof verification failed: %v", err)
				}
				assigningInputs.Zkevm = fullZkEvm
				assigningInputs.ZkevmWizardProof = proof
			}

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
		serializedProof = dummy.MakeProof(&setup, funcInput.SumAsField(), mockCircuitID)
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

// buildMockZkevmInputs converts transaction data into the Inputs struct
// expected by MockZkevmArithCols. This is used when no real execution traces
// or Shomei traces are available (test/mock mode for BadPrecompile/TooManyLogs).
func buildMockZkevmInputs(tx *ethtypes.Transaction, fromAddr linTypes.EthAddress, req *Request) invalidityPI.Inputs {
	txHash := ethereum.GetTxHash(tx)

	var txHashLimbs [16]field.Element
	for i := 0; i < 16; i++ {
		txHashLimbs[i] = field.NewElement(uint64(txHash[30-2*i])<<8 | uint64(txHash[31-2*i]))
	}

	var fromLimbs [10]field.Element
	for i := 0; i < 10; i++ {
		fromLimbs[i] = field.NewElement(uint64(fromAddr[18-2*i])<<8 | uint64(fromAddr[19-2*i]))
	}

	var stateRootLimbs [8]field.Element
	for i := 0; i < 8; i++ {
		stateRootLimbs[i] = field.Element(req.ZkParentStateRootHash[i])
	}

	hasBadPrecompile := req.InvalidityType == invalidity.BadPrecompile
	numL2Logs := 0
	if req.InvalidityType == invalidity.TooManyLogs {
		numL2Logs = 20 // > MAX_L2_LOGS (16)
	}

	return invalidityPI.Inputs{
		FixedInputs: invalidityPI.FixedInputs{
			TxHashLimbs:    txHashLimbs,
			FromLimbs:      fromLimbs,
			StateRootLimbs: stateRootLimbs,
			ColSize:        16,
		},
		CaseInputs: invalidityPI.CaseInputs{
			HasBadPrecompile: hasBadPrecompile,
			NumL2Logs:        numL2Logs,
		},
	}
}
