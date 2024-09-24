//go:build !fuzzlight

package v1_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	v1 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1/test_utils"
	"github.com/consensys/linea-monorepo/prover/utils/types"

	"github.com/consensys/linea-monorepo/prover/backend/ethereum"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecode(t *testing.T) {

	testBlocks, _ := test_utils.TestBlocksAndBlobMaker(t)

	for i, rlpBlock := range testBlocks {
		t.Run(fmt.Sprintf("block-#%v", i), func(t *testing.T) {

			var block ethtypes.Block
			if err := rlp.Decode(bytes.NewReader(rlpBlock), &block); err != nil {
				t.Fatalf("could not decode test RLP block: %s", err.Error())
			}

			var (
				buf      = &bytes.Buffer{}
				expected = test_utils.DecodedBlockData{
					BlockHash: block.Hash(),
					Txs:       make([]ethtypes.Transaction, len(block.Transactions())),
					Timestamp: block.Time(),
				}
			)

			for i := range expected.Txs {
				expected.Txs[i] = *block.Transactions()[i]
			}

			if err := v1.EncodeBlockForCompression(&block, buf); err != nil {
				t.Fatalf("failed encoding the block: %s", err.Error())
			}

			var (
				encoded       = buf.Bytes()
				r             = bytes.NewReader(encoded)
				decoded, err  = test_utils.DecodeBlockFromUncompressed(r)
				size, errScan = v1.ScanBlockByteLen(encoded)
			)

			assert.NoError(t, errScan, "error scanning the payload length")
			assert.NotZero(t, size, "scanned a block size of zero")

			require.NoError(t, err)
			assert.Equal(t, expected.BlockHash, decoded.BlockHash)
			assert.Equal(t, expected.Timestamp, decoded.Timestamp)
			assert.Equal(t, len(expected.Txs), len(decoded.Txs))

			for i := range expected.Txs {
				checkSameTx(t, &expected.Txs[i], &decoded.Txs[i], decoded.Froms[i])
				if t.Failed() {
					return
				}
			}
		})
	}

}

func checkSameTx(t *testing.T, orig, decoded *ethtypes.Transaction, from common.Address) {
	assert.Equal(t, orig.To(), decoded.To())
	assert.Equal(t, orig.Nonce(), decoded.Nonce())
	assert.Equal(t, orig.Data(), decoded.Data())
	assert.Equal(t, orig.Value(), decoded.Value())
	assert.Equal(t, orig.Cost(), decoded.Cost())
	assert.Equal(t, ethereum.GetFrom(orig), types.EthAddress(from))
}

func TestPassRlpList(t *testing.T) {

	makeRlpSlice := func(n int) []byte {
		slice := make([]any, n)
		for i := range slice {
			// This is serialized in a single byte
			slice[i] = []any{}
		}

		b, err := rlp.EncodeToBytes(slice)
		if err != nil {
			utils.Panic("err = %v", err.Error())
		}

		return b
	}

	const (
		maxListSize = 1 << 12
	)

	for i := 0; i < maxListSize; i++ {
		var (
			length = i
			slice  = makeRlpSlice(length)
			r      = bytes.NewReader(slice)
		)

		if err := v1.PassRlpList(r); err != nil {
			t.Fatalf("failed for length: %v: %s", length, err.Error())
		}

		assert.Equal(t, 0, r.Len(), "the entire reader was not read (length = %v)", length)
	}
}

func TestVectorDecode(t *testing.T) {

	cases := []string{
		"000165c05627341299696b345fbbbdb4a5f55168ee397b58e339c572abd4239ba549dd69e274fe3b557e8fb62b89f4916b721be55ceb828dbd7302ef8205397084625900808462590080825208948d97689c9818892b700e27f316cc3e41e17fbeb9865af3107a400080c0",
	}

	postPad := [4]byte{}

	for _, c := range cases {
		b, err := hex.DecodeString(c)
		if err != nil {
			t.Fatal(err)
		}

		var (
			postPadded = append(b, postPad[:]...)
			r          = bytes.NewReader(b)
			_, errDec  = test_utils.DecodeBlockFromUncompressed(r)
			_, errScan = v1.ScanBlockByteLen(postPadded)
		)

		assert.NoError(t, errScan)
		assert.NoError(t, errDec)
		assert.Equal(t, 0, r.Len())
	}
}
