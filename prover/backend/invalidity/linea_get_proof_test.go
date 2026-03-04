package invalidity_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/consensys/linea-monorepo/prover/backend/ethereum"
	backend "github.com/consensys/linea-monorepo/prover/backend/invalidity"
	"github.com/consensys/linea-monorepo/prover/crypto/state-management/smt_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeAccountTrieInputs_FromRequestFile(t *testing.T) {
	fileBytes, err := os.ReadFile("testdata/5-1-getZkInvalidityProof-v2.json")
	require.NoError(t, err)

	var req backend.Request
	err = json.Unmarshal(fileBytes, &req)
	require.NoError(t, err)

	tx, err := ethereum.RlpDecodeWithSignature(req.RlpEncodedTx)
	require.NoError(t, err)

	fromAddress := ethereum.GetFrom(tx)

	inputs, addr, topRoot, err := backend.DecodeAccountTrieInputs(req.AccountMerkleProof)
	require.NoError(t, err)

	t.Logf("fromAddress (from RLP)  = %s", fromAddress.Hex())
	t.Logf("subRoot (inputs.Root)  = %s", types.KoalaOctuplet(inputs.SubRoot).Hex())
	t.Logf("topRoot (computed)     = %s", topRoot.Hex())
	t.Logf("zkParentStateRootHash  = %s", req.ZkParentStateRootHash.Hex())
	t.Logf("subRoot == zkParent?   %v", types.KoalaOctuplet(inputs.SubRoot) == req.ZkParentStateRootHash)

	assert.Equal(t, int64(0), inputs.Account.Nonce, "non-existing account nonce should be 0")
	assert.Equal(t, "0", inputs.Account.Balance.String(), "non-existing account balance should be 0")

	assert.Equal(t, smt_koalabear.DefaultDepth, len(inputs.LeafOpeningMinus.Proof.Siblings),
		"minus proof should have %d siblings", smt_koalabear.DefaultDepth)
	assert.Equal(t, 31, inputs.LeafOpeningMinus.Proof.Path, "leftLeafIndex should be 31")

	assert.Equal(t, smt_koalabear.DefaultDepth, len(inputs.LeafOpeningPlus.Proof.Siblings),
		"plus proof should have %d siblings", smt_koalabear.DefaultDepth)
	assert.Equal(t, 1, inputs.LeafOpeningPlus.Proof.Path, "rightLeafIndex should be 1")

	// Verify Leaf == Hash(LeafOpening) for minus and plus
	assert.Equal(t, field.Octuplet(inputs.LeafOpeningMinus.LeafOpening.Hash()), inputs.LeafOpeningMinus.Leaf,
		"minus leaf hash should match Hash(LeafOpening)")
	assert.Equal(t, field.Octuplet(inputs.LeafOpeningPlus.LeafOpening.Hash()), inputs.LeafOpeningPlus.Leaf,
		"plus leaf hash should match Hash(LeafOpening)")

	// Recover root from both proofs and verify they match the subRoot
	recoveredMinus, err := smt_koalabear.RecoverRoot(&inputs.LeafOpeningMinus.Proof, inputs.LeafOpeningMinus.Leaf)
	require.NoError(t, err)
	assert.Equal(t, field.Octuplet(inputs.SubRoot), recoveredMinus,
		"minus proof should recover the subRoot")

	recoveredPlus, err := smt_koalabear.RecoverRoot(&inputs.LeafOpeningPlus.Proof, inputs.LeafOpeningPlus.Leaf)
	require.NoError(t, err)
	assert.Equal(t, field.Octuplet(inputs.SubRoot), recoveredPlus,
		"plus proof should recover the subRoot")

	// Verify Hash(address) is between hKey(left) and hKey(right)
	hKey := backend.HashAddress(addr)
	assert.Equal(t, -1, inputs.LeafOpeningMinus.LeafOpening.HKey.Cmp(hKey),
		"hKey(minus) should be less than Hash(address)")
	assert.Equal(t, -1, hKey.Cmp(inputs.LeafOpeningPlus.LeafOpening.HKey),
		"Hash(address) should be less than hKey(plus)")

	// Verify topRoot matches ZkParentStateRootHash from the request
	assert.Equal(t, req.ZkParentStateRootHash, topRoot,
		"topRoot should match ZkParentStateRootHash from request")

	// Verify fromAddress (form RLP) is the same as the key (from json/shomei)
	assert.Equal(t, fromAddress, addr,
		"fromAddress should be the same as the key")

	t.Logf("Decoded non-existing: key=%x, minusIdx=%d, plusIdx=%d, topRoot=%x",
		addr, inputs.LeafOpeningMinus.Proof.Path, inputs.LeafOpeningPlus.Proof.Path, topRoot)
}

func TestDecodeAndRecoverRoot_ExistingAccount(t *testing.T) {
	fileBytes, err := os.ReadFile("testdata/proof.json")
	require.NoError(t, err)

	var response struct {
		AccountProof json.RawMessage `json:"accountProof"`
	}
	err = json.Unmarshal(fileBytes, &response)
	require.NoError(t, err)

	inputs, addr, _, err := backend.DecodeAccountTrieInputs(response.AccountProof)
	require.NoError(t, err)
	require.True(t, inputs.AccountExists)

	// Verify hKey == Hash(address)
	hKey := backend.HashAddress(addr)
	assert.Equal(t, hKey, inputs.LeafOpening.LeafOpening.HKey,
		"hKey in leaf opening should equal Hash(address)")

	// Verify Leaf == Hash(LeafOpening)
	assert.Equal(t, field.Octuplet(inputs.LeafOpening.LeafOpening.Hash()), inputs.LeafOpening.Leaf,
		"leaf hash should match Hash(LeafOpening)")

	// Recover root and verify it matches the subRoot from node[0]
	recoveredRoot, err := smt_koalabear.RecoverRoot(&inputs.LeafOpening.Proof, inputs.LeafOpening.Leaf)
	require.NoError(t, err)
	assert.Equal(t, field.Octuplet(inputs.SubRoot), recoveredRoot,
		"recovered root should match subRoot from JSON")

	t.Logf("Root match: %x", inputs.SubRoot)
}
