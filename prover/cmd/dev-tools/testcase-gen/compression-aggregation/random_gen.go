package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"

	"github.com/consensys/linea-monorepo/prover/backend/aggregation"
	"github.com/consensys/linea-monorepo/prover/backend/blobsubmission"
	"github.com/consensys/linea-monorepo/prover/backend/invalidity"
	circInvalidity "github.com/consensys/linea-monorepo/prover/circuits/invalidity"
	"github.com/consensys/linea-monorepo/prover/utils"
	linTypes "github.com/consensys/linea-monorepo/prover/utils/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// RandDataGen generates random data for the smart-contract
type RandDataGen struct {
	*rand.Rand
}

// Rand request spec. This is a union struct for either blob submissions or
// aggregations requests.
type RandGenSpec struct {
	// Optional, if set indicates that this should generate a blob submission
	BlobSubmissionSpec BlobSubmissionSpec `json:"blobSubmissionSpec"`
	// Optional, if set indicates that this should generate a proof aggregation
	AggregationSpec AggregationSpec `json:"aggregationSpec"`
	// Optional, if set indicates that this should generate an invalidity proof
	InvalidityProofSpec InvalidityProofSpec `json:"invalidityProofSpec"`
	// Optional, if set specifies the dynamic chain configuration to use
	DynamicChainConfigurationSpec DynamicChainConfigurationSpec `json:"dynamicChainConfigurationSpec"`
}

// DynamicChainConfiguration spec, specifies the chain configuration parameters
type DynamicChainConfigurationSpec struct {
	ChainID              uint64 `json:"chainID"`
	BaseFee              uint64 `json:"baseFee"`
	CoinBase             string `json:"coinBase"`
	L2MessageServiceAddr string `json:"l2MessageServiceAddr"`
}

// BlobSubmission spec, specifies how to generate a random blob submission
// prover output
type BlobSubmissionSpec struct {

	// If the IgnoreBefore flag is set to true, this tells the generator not
	// overwrite the provided parameters with the previous blob submission.
	// This is useful for generating invalid cases for the tests.
	IgnoreBefore bool `json:"ignoreBefore"`

	// L2 block numbers where the blob sequence starts and ends
	StartFromL2Block   int    `json:"startFromL2Block"`
	NumConflation      int    `json:"numConflation"`
	BlockPerConflation int    `json:"blockPerConflation"`
	ParentDataHash     string `json:"parentDataHash"`

	// Compressed data size
	CompressedDataSize int `json:"compressedDataSize"`
	// Parent shnarf
	ParentShnarf     string `json:"parentShnarf"`
	ParentZkRootHash string `json:"parentZkStateRootHash"`

	// isEip4844 enabled tells the generator to specify that the response it
	// using EIP-4844.
	Eip4844Enabled bool `json:"eip4844Enabled"`
}

// InvalidityProofSpec
type InvalidityProofSpec struct {
	FtxNumber                int      `json:"ftxNumber"`
	ChainID                  *big.Int `json:"chainID"`
	ExpectedBlockHeight      int      `json:"expectedBlockHeight"`
	LastFinalizedBlockNumber int      `json:"lastFinalizedBlockNumber"`
}

// Aggregation spec
type AggregationSpec struct {

	// If the IgnoreBefore flag is set to true, this tells the generator not
	// overwrite the provided parameters with the previous blob submission.
	// This is useful for generating invalid cases for the tests.
	IgnoreBefore bool `json:"ignoreBefore"`

	// Finalized shnarf
	FinalShnarf                  string `json:"finalizedShnarf"`
	ParentAggregationFinalShnarf string `json:"parentAggregationFinalShnarf"`

	// Parent data hash and the list of data hashes to be finalized
	DataHashes     []string `json:"dataHashes"`
	DataParentHash string   `json:"dataParentHash"`

	// Parent zk root hash of the state over which we want to finalize. In 0x
	// prefixed hexstring.
	ParentStateRootHash string `json:"parentStateRootHash"`

	// Timestamp of the last already finalized L2 block
	LastFinalizedTimestamp uint `json:"lastFinalizedTimestamp"`

	// Timestamp of the last L2 block to be finalized
	FinalTimestamp uint `json:"finalTimestamp"`

	// Rolling hash of the L1 messages received on L2 so far for the state to be
	// currently finalized. In 0x prefixed Hexstring
	L1RollingHash              string `json:"l1RollingHash"`
	LastFinalizedL1RollingHash string `json:"lastFinalizedL1RollingHash"`

	// Number of L1 messages received on L2 so far for the state to be
	// currently finalized. Messaging Feedback loop messaging number -
	// part of public input
	L1RollingHashMessageNumber              uint `json:"l1RollingHashMessageNumber"`
	LastFinalizedL1RollingHashMessageNumber uint `json:"lastFinalizedL1RollingHashMessageNumber"`

	// L2 Merkle roots for L2 -> L1 messages - The prover will be calculating
	// Merkle roots for trees of a set depth
	HowManyL2Msgs uint `json:"howManyL2Msgs"`

	// The depth of the Merkle tree for the Merkle roots being anchored. This is
	// not used as part of the public inputs but is nonetheless helpful on the
	// contracts
	L2MsgTreeDepth uint `json:"treeDepth"`

	// Bytes array indicating which L2 blocks have “MessageSent” events. The
	// index starting from 1 + currentL2BlockNumber. If the value contains 1,
	// it means that that block has events
	L2MessagingBlocksOffsets string `json:"l2MessagingBlocksOffsets"`

	// Final block number
	LastFinalizedBlockNumber uint `json:"lastFinalizedBlockNumber"`
	FinalBlockNumber         uint `json:"finalBlockNumber"`

	// List of the invalidity proof responses
	InvalidityProofs []*invalidity.Response `json:"invalidityProofs"`

	// ParentFtxRollingHash is the stream hash of the parent aggregation
	ParentAggregationFtxRollingHash string `json:"parentAggregationFtxRollingHash"`
	ParentAggregationFtxNumber      int    `json:"parentAggregationFtxNumber"`

	FinalFtxRollingHash string `json:"finalFtxRollingHash"`
	FinalFtxNumber      uint   `json:"finalFtxNumber"`

	// Filtered addresses for address filter
	FilteredAddresses []string `json:"filteredAddresses"`
}

// Generates a random request file for a blob submission
func RandBlobSubmission(rng *rand.Rand, spec BlobSubmissionSpec) *blobsubmission.Request {
	rdg := RandDataGen{rng}
	return &blobsubmission.Request{
		CompressedData: rdg.Base64RandLen(spec.CompressedDataSize, -1),
		ConflationOrder: struct {
			StartingBlockNumber int   "json:\"startingBlockNumber\""
			UpperBoundaries     []int "json:\"upperBoundaries\""
		}{
			StartingBlockNumber: spec.StartFromL2Block,
			UpperBoundaries: rdg.AscendingSeq(
				spec.StartFromL2Block,
				spec.NumConflation,
				spec.BlockPerConflation,
			),
		},
		PrevShnarf:          spec.ParentShnarf,
		ParentStateRootHash: spec.ParentZkRootHash,
		FinalStateRootHash:  rdg.HexString(32),
		DataParentHash:      spec.ParentDataHash,
		Eip4844Enabled:      spec.Eip4844Enabled,
	}
}

// Generates a random request file for an aggregation request
func RandAggregation(rng *rand.Rand, spec AggregationSpec) *aggregation.CollectedFields {

	// Convert filtered addresses from spec
	filteredAddrs := make([]linTypes.EthAddress, len(spec.FilteredAddresses))
	for i, addrStr := range spec.FilteredAddresses {
		filteredAddrs[i] = linTypes.EthAddress(common.HexToAddress(addrStr))
	}

	cf := &aggregation.CollectedFields{
		ParentAggregationFinalShnarf:            spec.ParentAggregationFinalShnarf,
		FinalShnarf:                             spec.FinalShnarf,
		ParentStateRootHash:                     linTypes.HexToKoalabearOctupletLoose(spec.ParentStateRootHash),
		DataHashes:                              spec.DataHashes,
		DataParentHash:                          spec.DataParentHash,
		ParentAggregationLastBlockTimestamp:     spec.LastFinalizedTimestamp,
		FinalTimestamp:                          spec.FinalTimestamp,
		LastFinalizedL1RollingHash:              spec.LastFinalizedL1RollingHash,
		L1RollingHash:                           spec.L1RollingHash,
		LastFinalizedL1RollingHashMessageNumber: spec.LastFinalizedL1RollingHashMessageNumber,
		L1RollingHashMessageNumber:              spec.L1RollingHashMessageNumber,
		L2MsgTreeDepth:                          spec.L2MsgTreeDepth,
		HowManyL2Msgs:                           spec.HowManyL2Msgs,
		L2MessagingBlocksOffsets:                string(spec.L2MessagingBlocksOffsets),
		LastFinalizedBlockNumber:                spec.LastFinalizedBlockNumber,
		FinalBlockNumber:                        spec.FinalBlockNumber,
		LastFinalizedFtxRollingHash:             spec.ParentAggregationFtxRollingHash,
		LastFinalizedFtxNumber:                  uint(spec.ParentAggregationFtxNumber),
		// By default the final stream hash is the same as the parent. We will
		// overwrite it later if there is an invalidity proof in the spec.
		FinalFtxRollingHash: spec.ParentAggregationFtxRollingHash,
		FinalFtxNumber:      uint(spec.ParentAggregationFtxNumber),
		FilteredAddresses:   filteredAddrs,
	}

	if len(spec.InvalidityProofs) > 0 {
		invalidityProof := spec.InvalidityProofs[len(spec.InvalidityProofs)-1]
		cf.FinalFtxRollingHash = invalidityProof.FtxRollingHash.Hex()
		cf.FinalFtxNumber = uint(invalidityProof.ForcedTransactionNumber)
	}

	rdg := RandDataGen{rng}

	// Generate the L2 Messages Merkle root hashes. The right hand of the
	// addition is to ensure that we get a divCeil and the right-shift is to
	// divide by the number of elements in the tree;
	numL2MsgTrees := cf.HowManyL2Msgs + (1 << cf.L2MsgTreeDepth) - 1
	numL2MsgTrees >>= cf.L2MsgTreeDepth
	for _i := uint(0); _i < numL2MsgTrees; _i++ {
		cf.L2MsgRootHashes = append(cf.L2MsgRootHashes, rdg.HexString(32))
	}

	return cf
}

// RandonmInvalidTransaction returns a random invalid transaction from a random
// from address and writes fromAddress, tx and txHash to the spec file
func RandInvalidityProofRequest(rng *rand.Rand, spec *InvalidityProofSpec, specFile string) *invalidity.Request {

	var (
		signer         = types.NewLondonSigner(spec.ChainID)
		address        = common.HexToAddress("0xfeeddeadbeeffeeddeadbeeffeeddead01245678")
		TEST_ADDRESS_A = common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		TEST_HASH_F    = common.HexToHash("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
		TEST_HASH_A    = common.HexToHash("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	)

	// Generate a FIXED/deterministic private key for consistent testing
	deterministicSeed := fmt.Sprintf("fixed_test_seed_for_invalidity_proof_123456_%v", rng.Int63())
	hash := crypto.Keccak256([]byte(deterministicSeed))
	privKey, err := crypto.ToECDSA(hash)
	if err != nil {
		panic(err)
	}

	// Transaction nonce (will be different from account nonce to make it invalid)
	txNonce := rng.Uint64() % 100

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   spec.ChainID,
		Nonce:     txNonce,
		GasTipCap: big.NewInt(int64(112121212)),
		GasFeeCap: big.NewInt(int64(123543135)),
		Gas:       4531112,
		To:        &address,
		Value:     big.NewInt(int64(845315452)),
		Data:      common.Hex2Bytes("0xdeed8745a20f"),
		AccessList: types.AccessList{
			types.AccessTuple{Address: TEST_ADDRESS_A, StorageKeys: []common.Hash{TEST_HASH_A, TEST_HASH_F}},
		},
	})

	// Sign the transaction so that GetFrom can recover the sender
	signedTx, err := types.SignTx(tx, signer, privKey)
	if err != nil {
		panic(fmt.Sprintf("failed to sign transaction: %v", err))
	}

	// Get sender address
	fromAddress, err := types.Sender(signer, signedTx)
	if err != nil {
		panic(fmt.Sprintf("failed to get sender: %v", err))
	}

	// Encode the signed transaction
	rlpEncodedTxBytes, err := signedTx.MarshalBinary()
	if err != nil {
		panic(fmt.Sprintf("failed to marshal signed transaction: %v", err))
	}

	// Create mock account with different nonce (to make BadNonce case)
	// Account nonce is different from tx nonce, so tx is invalid
	accountNonce := int64(txNonce + 100) // Different from tx nonce
	account := invalidity.CreateMockEOAAccount(accountNonce, big.NewInt(1e18))
	accountMerkleProof := invalidity.CreateMockAccountMerkleProof(fromAddress, account)

	return &invalidity.Request{
		RlpEncodedTx:                     "0x" + common.Bytes2Hex(rlpEncodedTxBytes),
		ForcedTransactionNumber:          uint64(spec.FtxNumber),
		InvalidityType:                   circInvalidity.BadNonce,
		DeadlineBlockHeight:              uint64(spec.ExpectedBlockHeight),
		PrevFtxRollingHash:               linTypes.KoalaOctuplet{}, // Will be set by caller when chaining from previous proof
		SimulatedExecutionBlockNumber:    uint64(spec.LastFinalizedBlockNumber) + 1,
		SimulatedExecutionBlockTimestamp: 1700000000, // Mock timestamp
		AccountMerkleProof:               accountMerkleProof,
	}

}

// Returns a slice of random length in base64. If to is smaller than from, then
// return a slice of deterministic length `from`.
func (rdg *RandDataGen) Base64RandLen(from, to int) string {
	n := from
	if to > from {
		// #nosec G404 --we don't need a cryptographic RNG for testing purpose
		n = from + rand.Intn(to-from)
	}
	n = utils.DivCeil(n, 32) * 32 // round up to the next multiple of 32
	res := make([]byte, n)
	rdg.Read(res)

	// Also zeroize all the bytes that are at position multiples of 32. This
	// ensures that we will not have overflow when casting to the bls12 scalar
	// field.
	for i := 0; i < n; i += 32 {
		res[i] = 0
	}

	return base64.StdEncoding.EncodeToString(res)
}

// Returns a slice of random bytes in hexadecimal. The length is deterministic
// and the returned hex string is 0x prefixed.
func (rdg *RandDataGen) HexString(n int) string {
	res := make([]byte, n)
	rdg.Read(res)
	return "0x" + hex.EncodeToString(res)
}

// Returns a sequence of random ascending numbers with a given length, with a
// maximal increment between consecutive numbers.
func (rdg *RandDataGen) AscendingSeq(startFrom, nTerms, maxInc int) []int {
	ret := make([]int, nTerms)
	s0 := startFrom
	for i := range ret {
		// It needs to increase at least by one.
		incBy := 1
		if maxInc > incBy {
			incBy += rdg.Intn(maxInc - incBy)
		}
		// Increments s0 and assigns it
		s0 = s0 + incBy
		ret[i] = s0
	}

	return ret
}
