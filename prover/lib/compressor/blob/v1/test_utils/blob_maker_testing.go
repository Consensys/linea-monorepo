package test_utils

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/consensys/linea-monorepo/prover/backend/ethereum"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob"
	v1 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/consensys/compress/lzss"
	fr381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func GenTestBlob(t require.TestingT, maxNbBlocks int) []byte {
	testBlocks, bm := TestBlocksAndBlobMaker(t)

	// Compress blocks
	cptBlock := 0
	for i, block := range testBlocks {
		// get a random from 1 to 5

		bSize := RandIntn(5) + 1 // #nosec G404 -- false positive

		if cptBlock > bSize && i%3 == 0 {
			cptBlock = 0
			bm.StartNewBatch()
		}

		appended, err := bm.Write(block, false)
		if !appended || i == maxNbBlocks {
			assert.NoError(t, err, "append a valid block should not generate an error")
			break
		} else {
			cptBlock++
		}
	}
	return bm.Bytes()
}

func LoadTestBlocks(testDataDir string) (testBlocks [][]byte, err error) {
	entries, err := os.ReadDir(testDataDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		jsonString, err := os.ReadFile(filepath.Join(testDataDir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var proverInput execution.Request
		if err = json.Unmarshal(jsonString, &proverInput); err != nil {
			return nil, err
		}

		for _, block := range proverInput.Blocks() {
			var bb bytes.Buffer
			if err = block.EncodeRLP(&bb); err != nil {
				return nil, err
			}
			testBlocks = append(testBlocks, bb.Bytes())
		}
	}
	return testBlocks, nil
}

func RandIntn(n int) int {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return int(binary.BigEndian.Uint64(b[:]) % uint64(n))
}

func EmptyBlob(t require.TestingT) []byte {
	var headerB bytes.Buffer

	repoRoot, err := blob.GetRepoRootPath()
	assert.NoError(t, err)
	// Init bm
	bm, err := v1.NewBlobMaker(1000, filepath.Join(repoRoot, "prover/lib/compressor/compressor_dict.bin"))
	assert.NoError(t, err)

	if _, err = bm.Header.WriteTo(&headerB); err != nil {
		panic(err)
	}

	compressor, err := lzss.NewCompressor(GetDict(t))
	assert.NoError(t, err)

	var bb bytes.Buffer
	if _, err = v1.PackAlign(&bb, headerB.Bytes(), fr381.Bits-1, v1.WithAdditionalInput(compressor.Bytes())); err != nil {
		panic(err)
	}
	return bb.Bytes()
}

func SingleBlockBlob(t require.TestingT) []byte {
	testBlocks, bm := TestBlocksAndBlobMaker(t)

	ok, err := bm.Write(testBlocks[0], false)
	assert.NoError(t, err)
	assert.True(t, ok)

	return bm.Bytes()
}

// TinyTwoBatchBlob produces a blob with two batches, each consisting of one block
func TinyTwoBatchBlob(t require.TestingT) []byte {

	testBlocks, bm := TestBlocksAndBlobMaker(t)

	for _, block := range testBlocks[:2] {
		ok, err := bm.Write(block, false)
		assert.NoError(t, err)
		assert.True(t, ok)
		bm.StartNewBatch()
	}

	return bm.Bytes()
}

// ConsecutiveBlobs (t, i, j, k) produces three blobs from the ranges [0:i], [i:i+j], [i+j:i+j+k] of test blocks
// each block is in its own batch
func ConsecutiveBlobs(t require.TestingT, n ...int) [][]byte {
	testBlocks, bm := TestBlocksAndBlobMaker(t)

	res := make([][]byte, 0, len(n))
	for _, n := range n {
		for j := 0; j < n; j++ {
			ok, err := bm.Write(testBlocks[j], false)
			assert.NoError(t, err)
			assert.True(t, ok)
			bm.StartNewBatch()
		}
		testBlocks = testBlocks[n:]
		res = append(res, bm.Bytes())
		bm.Reset()
	}

	return res
}

func TestBlocksAndBlobMaker(t require.TestingT) ([][]byte, *v1.BlobMaker) {
	repoRoot, err := blob.GetRepoRootPath()
	assert.NoError(t, err)
	testBlocks, err := LoadTestBlocks(filepath.Join(repoRoot, "testdata/prover-v2/prover-execution/requests"))
	assert.NoError(t, err)
	// Init bm
	bm, err := v1.NewBlobMaker(40000, filepath.Join(repoRoot, "prover/lib/compressor/compressor_dict.bin"))
	assert.NoError(t, err)
	return testBlocks, bm
}

// DecodedBlockData is a wrapper struct storing the different fields of a block
// that we deserialize when decoding an ethereum block.
type DecodedBlockData struct {
	// BlockHash stores the decoded block hash
	BlockHash common.Hash
	// Timestamp holds the Unix timestamp of the block in
	Timestamp uint64
	// Froms stores the list of the sender address of every transaction
	Froms []common.Address
	// Txs stores the list of the decoded transactions.
	Txs []types.Transaction
}

// DecodeBlockFromUncompressed inverts [EncodeBlockForCompression]. It is primarily meant for
// testing and ensuring the encoding is bijective.
func DecodeBlockFromUncompressed(r *bytes.Reader) (DecodedBlockData, error) {

	var (
		decNumTxs    uint16
		decTimestamp uint32
		blockHash    common.Hash
	)

	if err := binary.Read(r, binary.BigEndian, &decNumTxs); err != nil {
		return DecodedBlockData{}, fmt.Errorf("could not decode nb txs: %w", err)
	}

	if err := binary.Read(r, binary.BigEndian, &decTimestamp); err != nil {
		return DecodedBlockData{}, fmt.Errorf("could not decode timestamp: %w", err)
	}

	if _, err := r.Read(blockHash[:]); err != nil {
		return DecodedBlockData{}, fmt.Errorf("could not read the block hash: %w", err)
	}

	var (
		numTxs     = int(decNumTxs)
		decodedBlk = DecodedBlockData{
			Froms:     make([]common.Address, numTxs),
			Txs:       make([]types.Transaction, numTxs),
			Timestamp: uint64(decTimestamp),
			BlockHash: blockHash,
		}
	)

	for i := 0; i < int(decNumTxs); i++ {
		if err := DecodeTxFromUncompressed(r, &decodedBlk.Txs[i], &decodedBlk.Froms[i]); err != nil {
			return DecodedBlockData{}, fmt.Errorf("could not decode transaction #%v: %w", i, err)
		}
	}

	return decodedBlk, nil
}

func DecodeTxFromUncompressed(r *bytes.Reader, tx *types.Transaction, from *common.Address) (err error) {
	if _, err := r.Read(from[:]); err != nil {
		return fmt.Errorf("could not read from address: %w", err)
	}

	if err := ethereum.DecodeTxFromBytes(r, tx); err != nil {
		return fmt.Errorf("could not deserialize transaction")
	}

	return nil
}

func GetDict(t require.TestingT) []byte {
	dict, err := blob.GetDict()
	require.NoError(t, err)
	return dict
}
