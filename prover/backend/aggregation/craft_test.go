package aggregation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/consensys/compress/lzss"
	"github.com/consensys/gnark-crypto/ecc"
	fr381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/gnark/test"
	"github.com/consensys/linea-monorepo/prover/backend/blobdecompression"
	pi_interconnection "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/dictionary"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/encode"
	v1 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/constraints"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/stretchr/testify/assert"
)

func TestMiniTrees(t *testing.T) {

	cases := []struct {
		MsgHashes []string
		Res       []string
	}{
		{
			MsgHashes: []string{
				"0x0000000000000000000000000000000000000000000000000000000000000000",
				"0x1111111111111111111111111111111111111111111111111111111111111111",
				"0x2222222222222222222222222222222222222222222222222222222222222222",
				"0x3333333333333333333333333333333333333333333333333333333333333333",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
			},
			Res: []string{
				"0x97d2505cd0c868c753353628fbb1aacc52bba62ddebac0536256e1e8560d4f27",
			},
		},
		{
			MsgHashes: []string{
				"0x0000000000000000000000000000000000000000000000000000000000000000",
				"0x1111111111111111111111111111111111111111111111111111111111111111",
				"0x2222222222222222222222222222222222222222222222222222222222222222",
				"0x3333333333333333333333333333333333333333333333333333333333333333",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x5555555555555555555555555555555555555555555555555555555555555555",
				"0x6666666666666666666666666666666666666666666666666666666666666666",
			},
			Res: []string{
				"0x52b5853ebe75cdc639ba9ed15de287bb918b9a0aba00b7aba087de5ee5d0528d",
			},
		},
		{
			MsgHashes: []string{
				"0x0000000000000000000000000000000000000000000000000000000000000000",
				"0x1111111111111111111111111111111111111111111111111111111111111111",
				"0x2222222222222222222222222222222222222222222222222222222222222222",
				"0x3333333333333333333333333333333333333333333333333333333333333333",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x5555555555555555555555555555555555555555555555555555555555555555",
				"0x6666666666666666666666666666666666666666666666666666666666666666",
				"0x0000000000000000000000000000000000000000000000000000000000000000",
				"0x0000000000000000000000000000000000000000000000000000000000000000",
				"0x0000000000000000000000000000000000000000000000000000000000000000",
				"0x0000000000000000000000000000000000000000000000000000000000000000",
				"0x0000000000000000000000000000000000000000000000000000000000000000",
				"0x0000000000000000000000000000000000000000000000000000000000000000",
				"0x0000000000000000000000000000000000000000000000000000000000000000",
			},
			Res: []string{
				"0x52b5853ebe75cdc639ba9ed15de287bb918b9a0aba00b7aba087de5ee5d0528d",
			},
		},
		{
			MsgHashes: []string{
				"0x0000000000000000000000000000000000000000000000000000000000000000",
				"0x1111111111111111111111111111111111111111111111111111111111111111",
				"0x2222222222222222222222222222222222222222222222222222222222222222",
				"0x3333333333333333333333333333333333333333333333333333333333333333",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x5555555555555555555555555555555555555555555555555555555555555555",
				"0x6666666666666666666666666666666666666666666666666666666666666666",
				"0x0000000000000000000000000000000000000000000000000000000000000000",
				"0x0000000000000000000000000000000000000000000000000000000000000000",
				"0x0000000000000000000000000000000000000000000000000000000000000000",
				"0x0000000000000000000000000000000000000000000000000000000000000000",
				"0x0000000000000000000000000000000000000000000000000000000000000000",
				"0x0000000000000000000000000000000000000000000000000000000000000000",
				"0x0000000000000000000000000000000000000000000000000000000000000000",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
				"0x4444444444444444444444444444444444444444444444444444444444444444",
			},
			Res: []string{
				"0x8b25bcdfa0bc56e9e67d3db3c513aa605c8584ac450fb14d62d46cef0fba6f7d",
				"0xcb876e4686e714c06dd52157412e91a490483e9b43e477984c615e6e5dd44b29",
			},
		},
	}

	for i, testcase := range cases {
		res := PackInMiniTrees(testcase.MsgHashes)
		assert.Equal(t, testcase.Res, res, "for case %v", i)
	}
}

func TestL1OffsetBlocks(t *testing.T) {

	testcases := []struct {
		Inps []bool
		Outs string
	}{
		{
			Inps: []bool{true, true, false, false, false},
			Outs: "0x00010002",
		},
		{
			Inps: []bool{false, true, false, false, true, true},
			Outs: "0x000200050006",
		},
	}

	for i, c := range testcases {
		o := PackOffsets(c.Inps)
		oHex := utils.HexEncodeToString(o)
		assert.Equal(t, c.Outs, oHex, "for case %v", i)
	}

}

func TestCollectFields(t *testing.T) {
	reqDir := config.WithRequestDir{
		RequestsRootDir: "testdata",
	}
	cfg := config.Config{ // TODO fill relevant fields (for collectFields)
		Environment: "",
		Version:     "",
		LogLevel:    0,
		AssetsDir:   "",
		Controller:  config.Controller{},
		Execution: config.Execution{
			WithRequestDir:     reqDir,
			ProverMode:         "",
			CanRunFullLarge:    false,
			ConflatedTracesDir: "",
		},
		BlobDecompression: config.BlobDecompression{
			WithRequestDir: reqDir,
		},
		Aggregation: config.Aggregation{},
		PublicInputInterconnection: config.PublicInput{
			MaxNbDecompression: 2,
			MaxNbExecution:     5,
			MaxNbCircuits:      0,
			ExecutionMaxNbMsg:  1,
			L2MsgMerkleDepth:   5,
			L2MsgMaxNbMerkle:   2,
			MockKeccakWizard:   true,
		},
		Debug: struct {
			Profiling bool `mapstructure:"profiling"`
			Tracing   bool `mapstructure:"tracing"`
		}{},
		Layer2: struct {
			ChainID           uint           `mapstructure:"chain_id" validate:"required"`
			MsgSvcContractStr string         `mapstructure:"message_service_contract" validate:"required,eth_addr"`
			MsgSvcContract    common.Address `mapstructure:"-"`
		}{},
		TracesLimits:      config.TracesLimits{},
		TracesLimitsLarge: config.TracesLimits{},
	}

	cfg.Layer2.MsgSvcContract = common.HexToAddress("971e727e956690b9957be6d51ec16e73acac83a7")
	cfg.Layer2.ChainID = 0xe705
	cfg.PublicInputInterconnection.L2MsgServiceAddr = cfg.Layer2.MsgSvcContract
	cfg.PublicInputInterconnection.ChainID = uint64(cfg.Layer2.ChainID)

	var req Request
	require.NoError(t, utils.ReadFromJSON("testdata/4461000-4461233-getZkAggregatedProof.json", &req))
	cf, err := collectFields(&cfg, &req)
	require.NoError(t, err)

	piC, err := pi_interconnection.Compile(cfg.PublicInputInterconnection)
	require.NoError(t, err)

	assignment, err := piC.Assign(pi_interconnection.Request{
		Decompressions: cf.DecompressionPI,
		Executions:     cf.ExecutionPI,
		Aggregation:    cf.AggregationPublicInput(&cfg),
	})

	require.NoError(t, err)

	require.NoError(t, test.IsSolved(piC.Circuit, &assignment, ecc.BLS12_377.ScalarField()))
}

func TestFakeBlob(t *testing.T) {
	var resp blobdecompression.Response
	require.NoError(t, utils.ReadFromJSON("testdata/responses/4459440-4461292-getZkBlobCompressionProof.json", &resp))
	blob, err := base64.StdEncoding.DecodeString(resp.CompressedData)
	require.NoError(t, err)
	const dictPath = "/Users/arya/linea-monorepo/prover/lib/compressor/compressor_dict.bin"
	dictStore := dictionary.NewStore(dictPath)
	header, payload, _, err := v1.DecompressBlob(blob, dictStore)
	if err != nil {
		return
	}

	const startBatchI = 20
	dict, err := os.ReadFile(dictPath)
	require.NoError(t, err)
	dictSum, err := dictionary.Checksum(dict, 1)
	require.NoError(t, err)
	newHeader := v1.Header{
		BatchSizes: header.BatchSizes[startBatchI : startBatchI+3],
	}
	copy(newHeader.DictChecksum[:], dictSum)
	compressor, err := lzss.NewCompressor(dict)
	require.NoError(t, err)
	start := sum(header.BatchSizes[:startBatchI])
	for i := 0; i < 3; i++ {
		end := start + header.BatchSizes[i+startBatchI]
		_, err = compressor.Write(payload[start:end])
		require.NoError(t, err)
		start = end
	}
	var bb bytes.Buffer
	_, err = newHeader.WriteTo(&bb)
	require.NoError(t, err)

	var bb2 bytes.Buffer
	bb2.Grow(128 * 1024)

	_, err = encode.PackAlign(&bb2, bb.Bytes(), fr381.Bits-1, encode.WithAdditionalInput(compressor.Bytes()))
	require.NoError(t, err)

	//bb2.Write(make([]byte, bb2.Cap()-bb2.Len()))

	headerBack, payloadBack, _, err := v1.DecompressBlob(bb2.Bytes(), dictStore)
	require.NoError(t, err)
	require.Equal(t, headerBack.BatchSizes, header.BatchSizes[startBatchI:startBatchI+3])
	totalSize := sum(header.BatchSizes[startBatchI : startBatchI+3])
	require.NoError(t, test_utils.BytesEqual(payload[start:start+totalSize], payloadBack))

	resp.CompressedData = base64.StdEncoding.EncodeToString(bb2.Bytes())
	require.NoError(t, utils.WriteToJSON("testdata/responses/new-4461000-4461233-getZkBlobCompressionProof.json", &resp))
}

func sum[T constraints.Integer](s []T) T {
	var res T
	for i := range s {
		res += s[i]
	}
	return res
}

func TestPrettifyJson(t *testing.T) {
	const dir = "testdata/responses"
	files, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		f, err := os.Open(filepath.Join(dir, file.Name()))
		require.NoError(t, err)
		var o any
		require.NoError(t, json.NewDecoder(f).Decode(&o))
		require.NoError(t, f.Close())
		f, err = os.OpenFile(filepath.Join(dir, file.Name()), os.O_CREATE|os.O_WRONLY, 0600)
		require.NoError(t, err)
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		require.NoError(t, errors.Join(enc.Encode(o), f.Close()))
	}
}
