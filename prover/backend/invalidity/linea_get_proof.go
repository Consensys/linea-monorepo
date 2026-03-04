package invalidity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/accumulator"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/types"
)

const (
	treeDepth            = smt_koalabear.DefaultDepth       // 40
	proofRelatedNodesLen = treeDepth + 2                    // 42
	nextFreeNodeBytes    = 64                               // nextFreeNode uses 64-byte padded encoding
	subRootBytes         = 32                               // subSmtRoot is always 32 bytes
	metadataBytes        = nextFreeNodeBytes + subRootBytes // 96
	hashBytes            = 32                               // hKey/hVal are always 32 bytes
	siblingNodeBytes     = 64                               // left(32) + right(32) child hashes
)

// accountProof is the union of existing and non-existing account proof formats
// returned by linea_getProof.
//
// Existing account:  key + leafIndex + proof
// Non-existing account: key + leftLeafIndex + rightLeafIndex + leftProof + rightProof
type accountProof struct {
	Key types.EthAddress `json:"key"`

	// Existing account fields
	LeafIndex *int       `json:"leafIndex,omitempty"`
	Proof     *proofData `json:"proof,omitempty"`

	// Non-existing account fields
	LeftLeafIndex  *int       `json:"leftLeafIndex,omitempty"`
	RightLeafIndex *int       `json:"rightLeafIndex,omitempty"`
	LeftProof      *proofData `json:"leftProof,omitempty"`
	RightProof     *proofData `json:"rightProof,omitempty"`
}

type proofData struct {
	Value             string   `json:"value"`
	ProofRelatedNodes []string `json:"proofRelatedNodes"`
}

// DecodeAccountTrieInputs decodes a linea_getProof accountProof response
// into AccountTrieInputs, the account address, and the topRoot
// (= Poseidon2(nextFreeNode || subRoot)). Supports two formats:
//
//   - Existing account: leafIndex + proof  (single proof, like proof.json)
//   - Non-existing account: leftProof + rightProof + leftLeafIndex + rightLeafIndex
//     (leftProof = minus neighbor, rightProof = plus neighbor)
//     addr = key(from shomei)
func DecodeAccountTrieInputs(rawProof json.RawMessage) (invalidity.AccountTrieInputs, types.EthAddress, types.KoalaOctuplet, error) {
	var proof accountProof
	if err := json.Unmarshal(rawProof, &proof); err != nil {
		return invalidity.AccountTrieInputs{}, types.EthAddress{}, types.KoalaOctuplet{}, fmt.Errorf("parsing accountProof: %w", err)
	}

	if proof.Proof != nil {
		return decodeExistingAccount(proof)
	}
	if proof.LeftProof != nil && proof.RightProof != nil {
		return decodeNonExistingAccount(proof)
	}
	return invalidity.AccountTrieInputs{}, types.EthAddress{}, types.KoalaOctuplet{},
		fmt.Errorf("accountProof has neither 'proof' (existing) nor 'leftProof'/'rightProof' (non-existing)")
}

// decodeExistingAccount handles the single-proof format for accounts that exist in the trie.
func decodeExistingAccount(proof accountProof) (invalidity.AccountTrieInputs, types.EthAddress, types.KoalaOctuplet, error) {
	if proof.LeafIndex == nil {
		return invalidity.AccountTrieInputs{}, types.EthAddress{}, types.KoalaOctuplet{}, fmt.Errorf("existing account proof missing leafIndex")
	}

	subRoot, nextFreeNode, lo, err := decodeSingleProof(*proof.LeafIndex, proof.Proof)
	if err != nil {
		return invalidity.AccountTrieInputs{}, types.EthAddress{}, types.KoalaOctuplet{}, fmt.Errorf("decoding account proof for existing account: %w", err)
	}

	account, err := parseAccountValue(proof.Proof.Value)
	if err != nil {
		return invalidity.AccountTrieInputs{}, types.EthAddress{}, types.KoalaOctuplet{}, err
	}

	topRoot := invalidity.ComputeTopRoot(nextFreeNode, subRoot)
	inputs := invalidity.AccountTrieInputs{
		Account:          account,
		LeafOpening:      lo,
		LeafOpeningMinus: lo,
		LeafOpeningPlus:  lo,
		SubRoot:          field.Octuplet(subRoot),
		NextFreeNode:     nextFreeNode,
		TopRoot:          field.Octuplet(topRoot),
		AccountExists:    true,
	}
	return inputs, proof.Key, topRoot, nil
}

// decodeNonExistingAccount handles the left/right proof format for accounts
// that do not exist in the trie (non-membership proof).
func decodeNonExistingAccount(proof accountProof) (invalidity.AccountTrieInputs, types.EthAddress, types.KoalaOctuplet, error) {
	if proof.LeftLeafIndex == nil || proof.RightLeafIndex == nil {
		return invalidity.AccountTrieInputs{}, types.EthAddress{}, types.KoalaOctuplet{},
			fmt.Errorf("non-existing account proof missing leftLeafIndex or rightLeafIndex")
	}

	leftRoot, nextFreeNode, loMinus, err := decodeSingleProof(*proof.LeftLeafIndex, proof.LeftProof)
	if err != nil {
		return invalidity.AccountTrieInputs{}, types.EthAddress{}, types.KoalaOctuplet{}, fmt.Errorf("decoding leftProof: %w", err)
	}

	_, _, loPlus, err := decodeSingleProof(*proof.RightLeafIndex, proof.RightProof)
	if err != nil {
		return invalidity.AccountTrieInputs{}, types.EthAddress{}, types.KoalaOctuplet{}, fmt.Errorf("decoding rightProof: %w", err)
	}

	topRoot := invalidity.ComputeTopRoot(nextFreeNode, leftRoot)

	// For non-existing accounts, LeafOpening reuses the minus proof/leaf for (to pass the Proof check trivially)
	// but HKey must be Hash(targetAddress) so the unconditional address check
	// and the wrapping check (minus.HKey < HKey < plus.HKey) both pass.
	loTarget := loMinus
	loTarget.LeafOpening.HKey = HashAddress(proof.Key)

	inputs := invalidity.AccountTrieInputs{
		Account:          types.Account{Balance: big.NewInt(0)},
		LeafOpening:      loTarget,
		LeafOpeningMinus: loMinus,
		LeafOpeningPlus:  loPlus,
		SubRoot:          field.Octuplet(leftRoot),
		NextFreeNode:     nextFreeNode,
		TopRoot:          field.Octuplet(topRoot),
		AccountExists:    false,
	}
	return inputs, proof.Key, topRoot, nil
}

// decodeSingleProof decodes one proof (left, right, or existing) from
// a leafIndex and proofData into an invalidity.LeafOpening, the subRoot,
// and the nextFreeNode value from the metadata.
func decodeSingleProof(leafIndex int, pd *proofData) (types.KoalaOctuplet, int64, invalidity.LeafOpening, error) {
	var zero invalidity.LeafOpening
	if len(pd.ProofRelatedNodes) != proofRelatedNodesLen {
		return types.KoalaOctuplet{}, 0, zero, fmt.Errorf(
			"expected %d proofRelatedNodes, got %d", proofRelatedNodesLen, len(pd.ProofRelatedNodes),
		)
	}

	nodes := pd.ProofRelatedNodes

	// --- proofRelatedNodes[0]: metadata ---
	// Layout: nextFreeNode(64 bytes) | subSmtRoot(32 bytes) = 96 bytes
	metadataRaw, err := utils.HexDecodeString(nodes[0])
	if err != nil {
		return types.KoalaOctuplet{}, 0, zero, fmt.Errorf("decoding metadata node[0]: %w", err)
	}
	if len(metadataRaw) != metadataBytes {
		return types.KoalaOctuplet{}, 0, zero, fmt.Errorf(
			"metadata node[0]: expected %d bytes, got %d", metadataBytes, len(metadataRaw),
		)
	}

	nextFreeNode, _, err := types.ReadInt64On64Bytes(bytes.NewReader(metadataRaw[:nextFreeNodeBytes]))
	if err != nil {
		return types.KoalaOctuplet{}, 0, zero, fmt.Errorf("parsing nextFreeNode: %w", err)
	}

	var subRoot types.KoalaOctuplet
	if err := subRoot.SetBytes(metadataRaw[nextFreeNodeBytes:]); err != nil {
		return types.KoalaOctuplet{}, 0, zero, fmt.Errorf("parsing subRoot: %w", err)
	}

	// --- proofRelatedNodes[41]: target leaf opening ---
	leafOpening, err := parseLeafOpening(nodes[proofRelatedNodesLen-1])
	if err != nil {
		return types.KoalaOctuplet{}, 0, zero, fmt.Errorf(
			"parsing target leaf node[%d]: %w", proofRelatedNodesLen-1, err,
		)
	}
	leaf := field.Octuplet(leafOpening.Hash())

	// --- Build Merkle proof siblings ---
	// siblings[k] = Poseidon2Hash(proofRelatedNodes[40-k])
	//   k=0: sibling leaf at proofRelatedNodes[40] (variable size)
	//   k=1..39: internal nodes at proofRelatedNodes[39..1] (64 bytes each)
	siblings := make([]types.KoalaOctuplet, treeDepth)

	siblingLeafRaw, err := utils.HexDecodeString(nodes[treeDepth])
	if err != nil {
		return types.KoalaOctuplet{}, 0, zero, fmt.Errorf(
			"decoding sibling leaf node[%d]: %w", treeDepth, err,
		)
	}
	if isAllZeros(siblingLeafRaw) {
		siblings[0] = types.KoalaOctuplet{}
	} else {
		siblings[0], err = poseidon2HashBytes(siblingLeafRaw)
		if err != nil {
			return types.KoalaOctuplet{}, 0, zero, fmt.Errorf("hashing sibling leaf: %w", err)
		}
	}

	for k := 1; k < treeDepth; k++ {
		nodeIdx := treeDepth - k
		nodeRaw, err := utils.HexDecodeString(nodes[nodeIdx])
		if err != nil {
			return types.KoalaOctuplet{}, 0, zero, fmt.Errorf("decoding node[%d]: %w", nodeIdx, err)
		}
		if len(nodeRaw) != siblingNodeBytes {
			return types.KoalaOctuplet{}, 0, zero, fmt.Errorf(
				"node[%d]: expected %d bytes, got %d", nodeIdx, siblingNodeBytes, len(nodeRaw),
			)
		}
		siblings[k], err = poseidon2HashBytes(nodeRaw)
		if err != nil {
			return types.KoalaOctuplet{}, 0, zero, fmt.Errorf("hashing node[%d]: %w", nodeIdx, err)
		}
	}

	lo := invalidity.LeafOpening{
		LeafOpening: leafOpening,
		Leaf:        leaf,
		Proof: smt_koalabear.Proof{
			Path:     leafIndex,
			Siblings: siblings,
		},
	}

	return subRoot, nextFreeNode, lo, nil
}

// parseAccountValue decodes the hex-encoded Shomei account value (192 bytes)
// via AccountShomeiTraces.UnmarshalJSON.
func parseAccountValue(hexValue string) (types.Account, error) {
	var account types.AccountShomeiTraces
	quoted := []byte(`"` + hexValue + `"`)
	if err := account.UnmarshalJSON(quoted); err != nil {
		return types.Account{}, fmt.Errorf("parsing account value: %w", err)
	}
	return account.Account, nil
}

// parseLeafOpening decodes a hex-encoded leaf (192 bytes) into an accumulator.LeafOpening.
// Layout: prev(64) + next(64) + hKey(32) + hVal(32)
func parseLeafOpening(hexStr string) (accumulator.LeafOpening, error) {
	const leafBytes = 192 // prev(64) + next(64) + hKey(32) + hVal(32)
	const ptrWidth = 64

	raw, err := utils.HexDecodeString(hexStr)
	if err != nil {
		return accumulator.LeafOpening{}, fmt.Errorf("hex decode: %w", err)
	}
	if len(raw) != leafBytes {
		return accumulator.LeafOpening{}, fmt.Errorf("leaf: expected %d bytes, got %d", leafBytes, len(raw))
	}

	prev, _, err := types.ReadInt64On64Bytes(bytes.NewReader(raw[0:ptrWidth]))
	if err != nil {
		return accumulator.LeafOpening{}, fmt.Errorf("reading prev: %w", err)
	}

	next, _, err := types.ReadInt64On64Bytes(bytes.NewReader(raw[ptrWidth : 2*ptrWidth]))
	if err != nil {
		return accumulator.LeafOpening{}, fmt.Errorf("reading next: %w", err)
	}

	var hKey, hVal types.KoalaOctuplet
	if err := hKey.SetBytes(raw[2*ptrWidth : 2*ptrWidth+hashBytes]); err != nil {
		return accumulator.LeafOpening{}, fmt.Errorf("reading hKey: %w", err)
	}
	if err := hVal.SetBytes(raw[2*ptrWidth+hashBytes : 2*ptrWidth+2*hashBytes]); err != nil {
		return accumulator.LeafOpening{}, fmt.Errorf("reading hVal: %w", err)
	}

	return accumulator.LeafOpening{
		Prev: prev,
		Next: next,
		HKey: hKey,
		HVal: hVal,
	}, nil
}

// poseidon2HashBytes computes the Poseidon2 hash of a byte slice,
// interpreting the bytes as a sequence of KoalaBear field elements.
func poseidon2HashBytes(data []byte) (types.KoalaOctuplet, error) {
	digest := poseidon2_koalabear.HashBytes(data)
	var result types.KoalaOctuplet
	if err := result.SetBytes(digest); err != nil {
		return types.KoalaOctuplet{}, fmt.Errorf("poseidon2 hash result: %w", err)
	}
	return result, nil
}

func isAllZeros(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}
