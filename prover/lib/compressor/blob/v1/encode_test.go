package v1_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"os"
	"path"
	"testing"

	"github.com/consensys/linea-monorepo/prover/backend/blobdecompression"
	v1 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeTx(t *testing.T) {

	var (
		privKey, _ = ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
		chainID    = big.NewInt(51)
		signer     = types.NewLondonSigner(chainID)
	)

	testTx := []struct {
		Name string
		Tx   types.TxData
	}{
		{
			Name: "contract-deployment-legacy",
			Tx: &types.LegacyTx{
				Nonce:    3,
				GasPrice: big.NewInt(10002),
				Gas:      7000007,
				To:       nil,
				Value:    big.NewInt(66666666),
				Data:     hexutil.MustDecode("0xdeadbeafbeefbeef12345689"),
			},
		},
		{
			Name: "contract-tx-legacy",
			Tx: &types.LegacyTx{
				Nonce:    3,
				GasPrice: big.NewInt(10002),
				Gas:      7000007,
				To:       &common.Address{12, 24},
				Value:    big.NewInt(66666666),
				Data:     hexutil.MustDecode("0xdeadbeafbeefbeef12345689"),
			},
		},
		{
			Name: "payment-legacy",
			Tx: &types.LegacyTx{
				Nonce:    3,
				GasPrice: big.NewInt(10002),
				Gas:      7000007,
				To:       &common.Address{12, 24},
				Value:    big.NewInt(66666666),
				Data:     nil,
			},
		},
		{
			Name: "contract-deployment-legacy",
			Tx: &types.DynamicFeeTx{
				Nonce:     3,
				GasTipCap: big.NewInt(10002),
				GasFeeCap: big.NewInt(33333),
				Gas:       7000007,
				To:        nil,
				Value:     big.NewInt(66666666),
				Data:      hexutil.MustDecode("0xdeadbeafbeefbeef12345689"),
				ChainID:   chainID,
			},
		},
		{
			Name: "contract-tx-legacy",
			Tx: &types.DynamicFeeTx{
				Nonce:     3,
				GasTipCap: big.NewInt(10002),
				GasFeeCap: big.NewInt(33333),
				Gas:       7000007,
				To:        &common.Address{12, 24},
				Value:     big.NewInt(66666666),
				Data:      hexutil.MustDecode("0xdeadbeafbeefbeef12345689"),
				ChainID:   chainID,
			},
		},
		{
			Name: "payment-legacy",
			Tx: &types.DynamicFeeTx{
				Nonce:     3,
				GasTipCap: big.NewInt(10002),
				GasFeeCap: big.NewInt(33333),
				Gas:       7000007,
				To:        &common.Address{12, 24},
				Value:     big.NewInt(66666666),
				Data:      nil,
				ChainID:   chainID,
			},
		},
	}

	for _, tc := range testTx {

		t.Run(tc.Name, func(t *testing.T) {

			var (
				tx   = types.MustSignNewTx(privKey, signer, tc.Tx)
				buf  = &bytes.Buffer{}
				addr = &common.Address{}
			)

			if err := v1.EncodeTxForCompression(tx, buf); err != nil {
				t.Fatalf("could not encode the transaction")
			}

			var (
				data = buf.Bytes()
				r    = bytes.NewReader(data)
			)

			txData2, err := v1.DecodeTxFromUncompressed(r, addr)

			if err != nil {
				t.Fatalf("could not deserialize the transaction err=%v", err.Error())
			}

			tx2 := types.NewTx(txData2)

			assert.Equal(t, tx.To(), tx2.To(), "field `to` mismatches")

		})

	}

}

func TestEncodeDecodeFromResponse(t *testing.T) {

	var (
		testDir        = "../testdata/v1/prover-responses"
		testFiles, err = os.ReadDir(testDir)
	)

	if err != nil {
		t.Fatalf("can't read test files: %v", err)
	}

	for _, testFile := range testFiles {

		if testFile.IsDir() {
			continue
		}

		fName := testFile.Name()

		if fName == ".DS_Store" {
			continue
		}

		t.Run(fName, func(t *testing.T) {

			var (
				filePath = path.Join(testDir, fName)
				f, err   = os.Open(filePath)
			)

			require.NoErrorf(t, err, "could not open the test file path=%v err=%v", filePath, err)

			var response = &blobdecompression.Response{}
			if err = json.NewDecoder(f).Decode(response); err != nil {
				t.Fatalf("could not deserialize testfile path=%v err=%v", filePath, err)
			}

			data, err := base64.StdEncoding.DecodeString(response.CompressedData)
			if err != nil {
				t.Fatalf("could not deserialize the bsae64 decompression data err=%v", err)
			}

			batches, err := decompressBlob(data)
			if err != nil {
				t.Fatalf("could not decompress the blob: err=%v", err)
			}

			for batchI := range batches {
				for blockI := range batches[batchI] {

					r := bytes.NewReader(batches[batchI][blockI])
					_, err := v1.DecodeBlockFromUncompressed(r)

					if err != nil {
						t.Fatalf("could not decode block: %v", err)
					}

					// for txI, txData := range blockData.Txs {
					// 	tx := types.NewTx(txData)

					// 	if tx.To() == nil {
					// 		t.Logf("batch=%v block=%v tx=%v tx.to=%v", batchI, blockI, txI, tx.To())
					// 		t.Fail()
					// 	}

					// 	if tx.To() == (&common.Address{}) {
					// 		t.Logf("batch=%v block=%v tx=%v tx.to=%v", batchI, blockI, txI, tx.To())
					// 		t.Fail()
					// 	}
					// }

				}
			}
		})

	}

}
