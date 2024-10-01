package blob_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/dictionary"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/encode"
	v0 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v0"
	blobv1testing "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1/test_utils"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestGetVersion(t *testing.T) {
	_blob := blobv1testing.GenTestBlob(t, 1)
	assert.Equal(t, uint32(0x10000), uint32(0xffff)+uint32(blob.GetVersion(_blob)), "version should match the current one")
}

const dictPath = "../compressor_dict.bin"

// TODO generalize these tests so that they apply to v1 as well
// TODO though the diagnostic tools will have to be version dependent

// Once a version of the compressor is released, the encoding/compression scheme must not change
// So we should be able to take a blob from the wild, read it, re-(encode/compress) it and get
// the same blob back
func TestBlobRoundTripV0(t *testing.T) {
	dictStore := dictionary.NewStore()
	require.NoError(t, dictStore.Load(dictPath))
	blobData := readHexFile(t, "testdata/v0/sample-blob-01b9918c3f0ceb6a.hex")
	header, payload, blocksSerialized, err := v0.DecompressBlob(blobData, dictStore)
	require.NoError(t, err)
	bm, err := v0.NewBlobMaker(v0.MaxUncompressedBytes, "../compressor_dict.bin")
	require.NoError(t, err)
	for i := 0; i < header.NbBatches(); i++ {
		for j := 0; j < header.NbBlocksInBatch(i); j++ {
			dbd, err := v0.DecodeBlockFromUncompressed(bytes.NewReader(blocksSerialized[0]))
			assert.NoError(t, err)

			stdBlockRlp, err := rlp.EncodeToBytes(dbd.ToStd())

			ok, err := bm.Write(stdBlockRlp, false, encode.WithTxAddressGetter(encode.GetAddressFromR))
			assert.NoError(t, err)
			assert.True(t, ok)
			blocksSerialized = blocksSerialized[1:]
		}
		bm.StartNewBatch()
	}
	assert.Empty(t, blocksSerialized)
	//test_utils.BytesEqual(t, bm.Bytes(), blobData)
	// can't expect the compressor to be deterministic. But we can decompress the produced blob and check if the header and raw data are equal
	headerBack, payloadBack, _, err := v0.DecompressBlob(bm.Bytes(), dictStore)
	assert.NoError(t, err)

	//assert.True(t, header.Equals(headerBack))
	require.NoError(t, header.CheckEquality(headerBack))

	if err = test_utils.BytesEqual(payload, payloadBack); err != nil {
		var bytesEqualError *test_utils.BytesEqualError
		errors.As(err, &bytesEqualError)
		failure := bytesEqualError.Index

		var batchI, blockI, blockAbsI, blockByteI int
		func() { // compute the above values for the failure point
			for batchI = 0; batchI < header.NbBatches(); batchI++ {
				for blockI = 0; blockI < header.NbBlocksInBatch(batchI); blockI++ {
					blockLength := header.BlockLength(batchI, blockI)
					if blockByteI+blockLength > failure {
						return
					}
					blockByteI += blockLength
					blockAbsI++
				}
			}
			t.Fatalf("failure point %d not found in payload of total length %d", failure, blockByteI)
		}()
		t.Error(err)
	}
}

// take a blob from the wild, decode and re-encode the blocks to see if they are the same
func TestBlockRoundTripV0(t *testing.T) {
	dictStore := dictionary.NewStore()
	require.NoError(t, dictStore.Load(dictPath))
	blobData := readHexFile(t, "testdata/v0/sample-blob-01b9918c3f0ceb6a.hex")
	_, _, blocksSerialized, err := v0.DecompressBlob(blobData, dictStore)
	require.NoError(t, err)

	txTypes := []string{"legacy", "access type", "dynamic"}

	var bb bytes.Buffer
	for blockI, block := range blocksSerialized {
		dbd, err := v0.DecodeBlockFromUncompressed(bytes.NewReader(block))
		assert.NoError(t, err)
		blockStdSerialized, err := rlp.EncodeToBytes(dbd.ToStd()) // blockStdSerialized simulates what the input of the compressor must have been
		assert.NoError(t, err)

		var blockStd types.Block
		assert.NoError(t, rlp.DecodeBytes(blockStdSerialized, &blockStd))

		bb.Reset()
		assert.NoError(t, v0.EncodeBlockForCompression(&blockStd, &bb, encode.WithTxAddressGetter(encode.GetAddressFromR)))

		err = test_utils.BytesEqual(block, bb.Bytes())
		assert.NoError(t, err, "at block #%d", blockI)
		if err != nil { // add more details to the error: find which transaction is causing the failure
			const transactionsStartAt = 64 / 8 // after the timestamp
			assert.Equal(t, block[:transactionsStartAt], bb.Bytes()[:transactionsStartAt], "timestamp mismatch")

			r := bytes.NewReader(block[transactionsStartAt:])
			rBack := bytes.NewReader(bb.Bytes()[transactionsStartAt:])

			for txI := 0; r.Len() != 0; txI++ {
				rNbUnreadBytesStart := r.Len()

				rlpDecoded, tp, err := v0.ReadTxAsRlp(r)
				require.NoError(t, err)
				rlpDecodedBack, tpBack, err := v0.ReadTxAsRlp(rBack)
				require.NoError(t, err)

				assert.Equal(t, tp, tpBack, "type mismatch on block #%d's tx #%d", blockI, txI)

				err = test_utils.SlicesEqual(rlpDecoded, rlpDecodedBack)
				assert.NoError(t, err, "at block #%d's tx #%d of type %s", blockI, txI, txTypes[tp])
				if err != nil { // print the whole transaction
					faultyTxEnd := len(block) - r.Len()
					faultyTxStart := len(block) - rNbUnreadBytesStart
					faultyTx := block[faultyTxStart:faultyTxEnd]
					t.Error("compressor-encoded transaction:", utils.HexEncodeToString(faultyTx))
					t.Error("compressor-encoded transaction (base64):", base64.StdEncoding.EncodeToString(faultyTx))
				}
			}
			assert.Equal(t, 0, rBack.Len())
		}
	}
}

// Take a decompressed transaction from a real blob, decode it into an ethereum object and back into a compressible transaction.
// the representation must not change
func TestTransactionRoundTripV0(t *testing.T) {
	test := func(b64 string) {
		tx, err := base64.StdEncoding.DecodeString(b64)
		require.NoError(t, err)
		var from common.Address
		txd, err := v0.DecodeTxFromUncompressed(bytes.NewReader(tx), &from)
		var bb bytes.Buffer
		require.NoError(t,
			v0.EncodeTxForCompression(encode.InjectFromAddressIntoR(txd, &from),
				&bb, encode.WithTxAddressGetter(encode.GetAddressFromR)),
		)

		fields, tp, err := v0.ReadTxAsRlp(bytes.NewReader(tx))
		require.NoError(t, err)
		fieldsBack, tpBack, err := v0.ReadTxAsRlp(bytes.NewReader(bb.Bytes()))
		require.NoError(t, err)

		assert.Equal(t, tp, tpBack, "transaction type mismatch")
		assert.NoError(t, test_utils.SlicesEqual(fields, fieldsBack))
	}

	// a contract creation transaction, so the To address is zero. extra care must be taken to get the same encoding.
	test("+QFLggJzhASHqwCDAopwlCGYwub6TkmgtT2QuiFqPThE/atAgIC5ASVggGBAUmAAgFRh//8ZFpBVNIAVYQAbV2AAgP1bUGD7gGEAKmAAOWAA8/5ggGBAUjSAFWAPV2AAgP1bUGAENhBgMldgADVg4ByAYwxVaZwUYDdXgGO0kATpFGBbV1tgAID9W2AAVGBEkGH//xaBVltgQFFh//+QkRaBUmAgAWBAUYCRA5DzW2BhYGNWWwBbYACAVGABkZCBkGB6kISQYf//FmCWVluSUGEBAAqBVIFh//8CGRaQg2H//xYCF5BVUFZbYf//gYEWg4IWAZCAghEVYL5XY05Ie3Fg4BtgAFJgEWAEUmAkYAD9W1CSkVBQVv6iZGlwZnNYIhIgZmyH7FASaIFylaTKH8bjhZ+vJB843WiPFFE1lwkgAJJkc29sY0MACBIAMw==")
}

// an M-trip is of the form A⇾B⇾A⇾B. We take a transaction in its original RLP form, serialize it into the compressor format,
// parse it back into a standard ethereum tx object, and serialize it into the compressor format again. The compressor encodings must be equal.
func TestTransactionMTrip(t *testing.T) {
	test := func(tx *types.Transaction) {
	}

	in, err := os.Open("testdata/blocks/9979248.json")
	require.NoError(t, err)

	var obj map[string]any
	require.NoError(t, json.NewDecoder(in).Decode(&obj))

	blockRlp, err := base64.StdEncoding.DecodeString(obj["block"].(string))
	require.NoError(t, err)

	blockRlp, err = os.ReadFile("testdata/blocks/0000000000984570")
	require.NoError(t, err)

	bm, err := v0.NewBlobMaker(12000000, "../compressor_dict.bin")
	require.NoError(t, err)
	ok, err := bm.Write(blockRlp, false)
	assert.NoError(t, err)
	require.True(t, ok)

	var block types.Block
	require.NoError(t, rlp.DecodeBytes(blockRlp, &block))

	test(nil)
}

func readHexFile(t *testing.T, filename string) []byte {
	hex, err := os.ReadFile(filename)
	require.NoError(t, err)
	res, err := utils.HexDecodeString(string(hex))
	require.NoError(t, err)
	return res
}
