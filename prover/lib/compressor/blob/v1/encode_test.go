//go:build !fuzzlight

package v1_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/encode"
	encodeTesting "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/encode/test_utils"
	"testing"

	v1 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1/test_utils"
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

			var buf bytes.Buffer

			if err := v1.EncodeBlockForCompression(&block, &buf); err != nil {
				t.Fatalf("failed encoding the block: %s", err.Error())
			}

			encoded := buf.Bytes()
			r := bytes.NewReader(encoded)
			decoded, err := v1.DecodeBlockFromUncompressed(r)
			size, errScan := v1.ScanBlockByteLen(encoded)

			assert.NoError(t, errScan, "error scanning the payload length")
			assert.NotZero(t, size, "scanned a block size of zero")

			require.NoError(t, err)
			assert.Equal(t, block.Hash(), decoded.BlockHash)
			assert.Equal(t, block.Time(), decoded.Timestamp)
			assert.Equal(t, len(block.Transactions()), len(decoded.Txs))

			for i := range block.Transactions() {
				encodeTesting.CheckSameTx(t, block.Transactions()[i], ethtypes.NewTx(decoded.Txs[i]), decoded.Froms[i])
				if t.Failed() {
					return
				}
			}

			t.Log("attempting RLP serialization")

			encoded, err = rlp.EncodeToBytes(decoded.ToStd())
			assert.NoError(t, err)

			var blockBack ethtypes.Block
			assert.NoError(t, rlp.Decode(bytes.NewReader(encoded), &blockBack))

			assert.Equal(t, block.Hash(), blockBack.ParentHash())
			assert.Equal(t, block.Time(), blockBack.Time())
			assert.Equal(t, len(block.Transactions()), len(blockBack.Transactions()))

			for i := range block.Transactions() {
				tx := blockBack.Transactions()[i]
				encodeTesting.CheckSameTx(t, block.Transactions()[i], ethtypes.NewTx(decoded.Txs[i]), common.Address(encode.GetAddressFromR(tx)))
				if t.Failed() {
					return
				}
			}

		})
	}

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
			_, errDec  = v1.DecodeBlockFromUncompressed(r)
			_, errScan = v1.ScanBlockByteLen(postPadded)
		)

		assert.NoError(t, errScan)
		assert.NoError(t, errDec)
		assert.Equal(t, 0, r.Len())
	}
}
