package blob_test

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/dictionary"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/encode"
	v0 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v0"
	v1 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1"
	blobv1testing "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1/test_utils"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVersion(t *testing.T) {
	_blob := blobv1testing.GenTestBlob(t, 1)
	assert.Equal(t, uint32(0x10000), uint32(0xffff)+uint32(blob.GetVersion(_blob)), "version should match the current one")
}

const dictPath = "../compressor_dict.bin"

func TestAddToBlob(t *testing.T) {
	dictStore := dictionary.NewStore()
	require.NoError(t, dictStore.Load(dictPath))
	blobData := withNoError(t, os.ReadFile, "testdata/v0/sample-blob-01b9918c3f0ceb6a.bin")
	header, _, blocksSerialized, err := v0.DecompressBlob(blobData, dictStore)
	require.NoError(t, err)

	blobData = withNoError(t, os.ReadFile, "testdata/v0/sample-blob-0151eda71505187b5.bin")
	_, _, blocksSerializedNext, err := v0.DecompressBlob(blobData, dictStore)
	require.NoError(t, err)

	bm, err := v0.NewBlobMaker(v0.MaxUsableBytes, "../compressor_dict.bin")
	require.NoError(t, err)
	var ok bool
	writeBlock := func(blocks *[][]byte) {
		dbd, err := v0.DecodeBlockFromUncompressed(bytes.NewReader((*blocks)[0]))
		assert.NoError(t, err)

		stdBlockRlp, err := rlp.EncodeToBytes(dbd.ToStd())

		ok, err = bm.Write(stdBlockRlp, false, encode.WithTxAddressGetter(encode.GetAddressFromR))
		assert.NoError(t, err)

		*blocks = (*blocks)[1:]
	}

	for i := 0; i < header.NbBatches(); i++ {
		for j := 0; j < header.NbBlocksInBatch(i); j++ {
			writeBlock(&blocksSerialized)
			assert.True(t, ok)
		}
		bm.StartNewBatch()
	}
	assert.Empty(t, blocksSerialized)

	util0 := 100 * bm.Len() / v0.MaxUsableBytes

	require.NoError(t, err)
	for ok { // all in one batch
		writeBlock(&blocksSerializedNext)
	}

	util1 := 100 * bm.Len() / v0.MaxUsableBytes

	fmt.Printf("%d%%\n%d%%\n", util0, util1)
}

func withNoError[X, Y any](t *testing.T, f func(X) (Y, error), x X) Y {
	y, err := f(x)
	require.NoError(t, err)
	return y
}

func TestDecompressBlob(t *testing.T) {
	store := dictionary.NewStore("../compressor_dict.bin")
	files := newRecursiveFolderIterator(t, "testdata")
	for files.hasNext() {
		f := files.next()
		if filepath.Ext(f.path) == ".bin" {
			t.Run(f.path, func(t *testing.T) {
				decompressed, err := blob.DecompressBlob(f.content, store)
				assert.NoError(t, err)
				t.Log("decompressed length", len(decompressed))

				// load decompressed blob as blocks
				var blocksSerialized [][]byte
				assert.NoError(t, rlp.DecodeBytes(decompressed, &blocksSerialized))
				t.Log("number of decoded blocks", len(blocksSerialized))
				for _, blockSerialized := range blocksSerialized {
					var b types.Block
					assert.NoError(t, rlp.DecodeBytes(blockSerialized, &b))
				}
			})
		}
	}
}

type dirEntryWithFullPath struct {
	path    string
	content os.DirEntry
}

// goes through all files in a directory and its subdirectories
type recursiveFolderIterator struct {
	toVisit []dirEntryWithFullPath
	t       *testing.T
	pathLen int
}

type file struct {
	content []byte
	path    string
}

func (i *recursiveFolderIterator) openDir(path string) {
	content, err := os.ReadDir(path)
	require.NoError(i.t, err)
	for _, c := range content {
		i.toVisit = append(i.toVisit, dirEntryWithFullPath{path: filepath.Join(path, c.Name()), content: c})
	}
}

func (i *recursiveFolderIterator) hasNext() bool {
	return i.peek() != nil
}

func (i *recursiveFolderIterator) next() *file {
	f := i.peek()
	if f != nil {
		i.toVisit = i.toVisit[:len(i.toVisit)-1]
	}
	return f
}

// counter-intuitively, peek does most of the work by ensuring the top of the stack is always a file
func (i *recursiveFolderIterator) peek() *file {
	for len(i.toVisit) != 0 {
		lastIndex := len(i.toVisit) - 1
		c := i.toVisit[lastIndex]
		if c.content.IsDir() {
			i.toVisit = i.toVisit[:lastIndex]
			i.openDir(c.path)
		} else {
			b, err := os.ReadFile(c.path)
			require.NoError(i.t, err)
			return &file{content: b, path: c.path[i.pathLen:]}
		}
	}
	return nil
}

func newRecursiveFolderIterator(t *testing.T, path string) *recursiveFolderIterator {
	res := recursiveFolderIterator{t: t, pathLen: len(path) + 1}
	res.openDir(path)
	return &res
}

func TestEncodeBlockWithType4Tx(t *testing.T) {
	// Create a blob with a single block, containing an EIP-7702, Set Code transaction.

	// blockRlpBase64 from https://explorer.devnet.linea.build/block/9128081
	// compressed using zlib
	const (
		blockRlpCompressedBase64 = "eNr6yXz6J1PjgiXaYeZsl2a0dIqk9Zwp6VrjxnNbat9dhon/VNM++EbdZ1gge8b3xb3jsVWrW7embztzReqykKv0lK4S4Q8Lnf46XJnsPoUBC1jAbGSeHnZvvu2FBxqfjvi5+DxSz79VunX6ItfHYWdOsa93XmA7/zk7R00l966idw8Fzk/LdXT/0X3HQcGh1b1G4tjE+bMXMHanRh7yiFh3emvmHE822asLZq41yan3537kvs9p8aTt3DsZGUY6aGju9pjYUm46haGZUaemJWPdi+0LGBkYSg0YfLwdGBhjN2OLGTnXf6EOMmvTK8REzj7Jiz9sL2W0yafbX+Bz6SZOD/EqhQ6YSvYFYS/kxaXPhC773+z6bNKBH3nRHg+kZ+asPcCYpL/1cfIWxYaGBYScuODxhiNOM/7IiMz6/eXEzPydKurrHJ+kzJ7ss2TqTOmKoB2hPxmddlxl+XGp6TkzX6vIl2wWBgjZzL5QYcrhFQ+353xYH7M3s3GBySvpOUk2+3waWho1/8gc+BH3I6bpOfMUD+8H2ZGy1afKRFwe+Jmuua82q6KYn3HBDHnL/RwZM74uMvvGu9rKN2yh96ENzW+CLt87qtyr13NQbIHkdktLzs5bnWv2il1+Pbki+f6L5BL/hwLSUy9XOWpfXsbGOD9a5kyqk7RFw5pNRQ4lG5fnvJ587Xl1yG2lH+s3TmRacmxB9L6D+eyi9jVPamdufSsgpyXF3Xfr1AQF5/kyWx9PPW3z/Edmc+fr6y2sYcwWTUEcU5Ra0j4dFz39hvXe6l8HHM6ZSV4/z53S0Mx4TmtB7WSNe24fL7yWklz5333uK/fY0vMn1M7kyO/5tub2Spe/Kgvkbz569PU596zK8zud7N5N1vPQ2LIroWBF9hrO6Mrmf3OXHTgACAAA///o8zd2"
		dictPath                 = "../dict/25-04-21.bin"
	)
	var block types.Block
	blockBytes := zlibDecompressBase64EncodedBytes(t, blockRlpCompressedBase64)
	require.NoError(t, rlp.DecodeBytes(blockBytes, &block))

	bm, err := v1.NewBlobMaker(127000, dictPath)
	require.NoError(t, err)

	ok, err := bm.Write(blockBytes, false)
	require.NoError(t, err)
	require.True(t, ok)

	blobBytes := bm.Bytes()

	decompressR, err := v1.DecompressBlob(blobBytes, dictionary.NewStore(dictPath))
	require.NoError(t, err)

	require.Equal(t, 1, len(decompressR.Blocks))

	decodedBlock, err := v1.DecodeBlockFromUncompressed(bytes.NewReader(decompressR.Blocks[0]))
	require.NoError(t, err)

	blockBack := decodedBlock.ToStd()
	require.Equal(t, 2, len(blockBack.Transactions()))

	tx := block.Transactions()[0]
	txBack := blockBack.Transactions()[0]
	require.Equal(t, types.SetCodeTxType, int(tx.Type()))
	require.Equal(t, types.SetCodeTxType, int(txBack.Type()))
}

func zlibDecompressBase64EncodedBytes(t *testing.T, b64 string) []byte {
	compressed, err := base64.StdEncoding.DecodeString(b64)
	require.NoError(t, err)

	zReader, err := zlib.NewReader(bytes.NewReader(compressed))
	require.NoError(t, err)

	var bb bytes.Buffer
	readBuf := make([]byte, 1024)

	for n := len(readBuf); n == len(readBuf); {
		n, err = zReader.Read(readBuf)
		if err != io.EOF {
			require.NoError(t, err)
		}
		bb.Write(readBuf[:n])
	}

	return bb.Bytes()
}
