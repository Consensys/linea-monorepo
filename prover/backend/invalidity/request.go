package invalidity

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/backend/execution/statemanager"
	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/sirupsen/logrus"
)

// Request file for a forced transaction attempted to be included in the current aggregation.
// The forcedTransactionNumbers from request files per aggregation should create a consecutive sequence.
type Request struct {
	// RLP encoding of the forced transaction (hex encoded with 0x prefix).
	RlpEncodedTx string `json:"ftxRLP"`

	// Transaction number assigned by L1 contract (decimal encoding)
	ForcedTransactionNumber uint64 `json:"ftxNumber"`

	// Previous FTX rolling hash, i.e. the FTX stream hash of the previous forced transaction.
	PrevFtxRollingHash types.Bls12377Fr `json:"prevFtxRollingHash"`

	// The block number deadline before which one expects to see the transaction (decimal encoding)
	DeadlineBlockHeight uint64 `json:"ftxBlockNumberDeadline"`

	// The type of invalidity for the forced transaction.
	// Valid values: BadNonce, BadBalance, BadPrecompile, TooManyLogs, FilteredAddressFrom, FilteredAddressTo
	InvalidityType invalidity.InvalidityType `json:"invalidityType"`

	// ZK parent state root hash
	ZkParentStateRootHash types.KoalaOctuplet `json:"zkParentStateRootHash"`

	// Path to conflated execution traces file (required for BadPrecompile, TooManyLogs cases)
	ConflatedExecutionTracesFile string `json:"conflatedExecutionTracesFile,omitempty"`

	// Account merkle proof from Shomei linea_getProof API (with proofRelatedNodes).
	// Required for BadNonce, BadBalance cases.
	AccountMerkleProof *ShomeiAccountProof `json:"accountMerkleProof,omitempty"`

	// ZK state merkle proof (full Shomei trace)
	// Required for BadPrecompile, TooManyLogs cases
	ZkStateMerkleProof [][]statemanager.DecodedTrace `json:"zkStateMerkleProof,omitempty"`
	// case of FilteredAddressFrom/FilteredAddressTo: accountMerkleProof=null, zkStateMerkleProof=null

	// Simulated execution block number (ParentAggregationLastBlockNumber + 1)
	SimulatedExecutionBlockNumber uint64 `json:"simulatedExecutionBlockNumber,omitempty"`

	// Simulated execution block timestamp
	SimulatedExecutionBlockTimestamp uint64 `json:"simulatedExecutionBlockTimestamp,omitempty"`
}

// AccountTrieInputs extracts the AccountTrieInputs from the AccountMerkleProof.
// Used for BadNonce and BadBalance cases.
func (req *Request) AccountTrieInputs() (invalidity.AccountTrieInputs, types.EthAddress, types.KoalaOctuplet, error) {
	if req.AccountMerkleProof == nil {
		return invalidity.AccountTrieInputs{}, types.EthAddress{}, types.KoalaOctuplet{},
			fmt.Errorf("accountMerkleProof is nil")
	}
	return DecodeAccountTrieInputs(*req.AccountMerkleProof)
}

// Validate checks that the required fields are present based on the InvalidityType.
// Both partial and full modes require the same trace inputs.
func (req *Request) Validate(proverMode config.ProverMode) error {

	logrus.Infof("Validating invalidity request: %+v, proverMode: %s", req, proverMode)

	if req.ZkParentStateRootHash == (types.KoalaOctuplet{}) {
		return fmt.Errorf("zkParentStateRootHash is required")
	}

	if req.SimulatedExecutionBlockNumber == 0 {
		return fmt.Errorf("simulatedExecutionBlockNumber is required")
	}
	if req.SimulatedExecutionBlockTimestamp == 0 {
		return fmt.Errorf("simulatedExecutionBlockTimestamp is required")
	}

	switch req.InvalidityType {

	case invalidity.BadNonce, invalidity.BadBalance:
		if proverMode != config.ProverModeDev {
			if err := req.validateAccountMerkleProof(); err != nil {
				return err
			}
		}

	case invalidity.BadPrecompile, invalidity.TooManyLogs:
		if proverMode != config.ProverModeDev {
			if req.ConflatedExecutionTracesFile == "" {
				return fmt.Errorf("conflatedExecutionTracesFile is required for %s invalidity type in %s mode", req.InvalidityType, proverMode)
			}
			if req.ZkStateMerkleProof == nil {
				return fmt.Errorf("zkStateMerkleProof is required for %s invalidity type in %s mode", req.InvalidityType, proverMode)
			}
		}
	case invalidity.FilteredAddressFrom, invalidity.FilteredAddressTo:
		// No additional fields required.

	default:
		return fmt.Errorf("unknown invalidity type: %s", req.InvalidityType)
	}

	return nil
}

// validateAccountMerkleProof performs comprehensive sanity checks on the
// decoded accountMerkleProof, covering both existing and non-existing accounts.
func (req *Request) validateAccountMerkleProof() error {
	inputs, addr, topRoot, err := req.AccountTrieInputs()
	if err != nil {
		return fmt.Errorf("accountMerkleProof decode error: %w", err)
	}

	// topRoot must match the request's zkParentStateRootHash
	if topRoot != req.ZkParentStateRootHash {
		return fmt.Errorf("topRoot mismatch: topRoot=%s, zkParentStateRootHash=%s",
			topRoot.Hex(), req.ZkParentStateRootHash.Hex())
	}

	hKey := HashAddress(addr)

	if inputs.AccountExists {
		return validateExistingAccount(inputs, hKey)
	}
	return validateNonExistingAccount(inputs, hKey)
}

func validateExistingAccount(inputs invalidity.AccountTrieInputs, hKey types.KoalaOctuplet) error {
	lo := inputs.ProofMinus

	if len(lo.Proof.Siblings) != smt_koalabear.DefaultDepth {
		return fmt.Errorf("proof siblings depth: expected %d, got %d",
			smt_koalabear.DefaultDepth, len(lo.Proof.Siblings))
	}

	if field.Octuplet(lo.LeafOpening.Hash()) != lo.Leaf {
		return fmt.Errorf("leaf hash mismatch: Hash(LeafOpening) != Leaf")
	}

	if lo.LeafOpening.HKey != hKey {
		return fmt.Errorf("hKey mismatch: leaf hKey=%s, Hash(address)=%s",
			lo.LeafOpening.HKey.Hex(), hKey.Hex())
	}

	recovered, err := smt_koalabear.RecoverRoot(&lo.Proof, lo.Leaf)
	if err != nil {
		return fmt.Errorf("recovering root from proof: %w", err)
	}
	if recovered != field.Octuplet(inputs.SubRoot) {
		return fmt.Errorf("recovered root mismatch: recovered=%v, subRoot=%v", recovered, inputs.SubRoot)
	}

	return nil
}

func validateNonExistingAccount(inputs invalidity.AccountTrieInputs, hKey types.KoalaOctuplet) error {
	minus := inputs.ProofMinus
	plus := inputs.ProofPlus

	// Both proofs must have correct depth
	if len(minus.Proof.Siblings) != smt_koalabear.DefaultDepth {
		return fmt.Errorf("minus proof siblings depth: expected %d, got %d",
			smt_koalabear.DefaultDepth, len(minus.Proof.Siblings))
	}
	if len(plus.Proof.Siblings) != smt_koalabear.DefaultDepth {
		return fmt.Errorf("plus proof siblings depth: expected %d, got %d",
			smt_koalabear.DefaultDepth, len(plus.Proof.Siblings))
	}

	// Leaf must equal Hash(LeafOpening) for both neighbors
	if field.Octuplet(minus.LeafOpening.Hash()) != minus.Leaf {
		return fmt.Errorf("minus leaf hash mismatch: Hash(LeafOpening) != Leaf")
	}
	if field.Octuplet(plus.LeafOpening.Hash()) != plus.Leaf {
		return fmt.Errorf("plus leaf hash mismatch: Hash(LeafOpening) != Leaf")
	}

	// Both proofs must recover the same subRoot
	recoveredMinus, err := smt_koalabear.RecoverRoot(&minus.Proof, minus.Leaf)
	if err != nil {
		return fmt.Errorf("recovering root from minus proof: %w", err)
	}
	if recoveredMinus != field.Octuplet(inputs.SubRoot) {
		return fmt.Errorf("minus recovered root mismatch: recovered=%v, subRoot=%v", recoveredMinus, inputs.SubRoot)
	}

	recoveredPlus, err := smt_koalabear.RecoverRoot(&plus.Proof, plus.Leaf)
	if err != nil {
		return fmt.Errorf("recovering root from plus proof: %w", err)
	}
	if recoveredPlus != field.Octuplet(inputs.SubRoot) {
		return fmt.Errorf("plus recovered root mismatch: recovered=%v, subRoot=%v", recoveredPlus, inputs.SubRoot)
	}

	// Hash(address) must be between hKey(minus) and hKey(plus)
	if minus.LeafOpening.HKey.Cmp(hKey) >= 0 {
		return fmt.Errorf("non-existing account: hKey(minus)=%s >= Hash(address)=%s",
			minus.LeafOpening.HKey.Hex(), hKey.Hex())
	}
	if hKey.Cmp(plus.LeafOpening.HKey) >= 0 {
		return fmt.Errorf("non-existing account: Hash(address)=%s >= hKey(plus)=%s",
			hKey.Hex(), plus.LeafOpening.HKey.Hex())
	}

	return nil
}
