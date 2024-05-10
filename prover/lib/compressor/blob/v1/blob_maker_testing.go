package v1

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/consensys/compress/lzss"
	fr381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/zkevm-monorepo/prover/backend/execution"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assumes that current working directory is within the repo
func getRepoRootPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	const repoName = "zkevm-monorepo"
	i := strings.LastIndex(wd, repoName)
	if i == -1 {
		return "", errors.New("could not find repo root")
	}
	i += len(repoName)
	return wd[:i], nil
}

func LoadDict(t require.TestingT) []byte {
	repoRoot, err := getRepoRootPath()
	assert.NoError(t, err)
	dictPath := filepath.Join(repoRoot, "prover/lib/compressor/compressor_dict.bin")
	dict, err := os.ReadFile(dictPath)
	assert.NoError(t, err)
	return dict
}

func GenTestBlob(t require.TestingT, maxNbBlocks int) []byte {
	assert := require.New(t) // taken from TestCompressorWithBatches
	repoRoot, err := getRepoRootPath()
	assert.NoError(err)
	testBlocks, err := loadTestBlocks(filepath.Join(repoRoot, "testdata/prover-v2/prover-execution/requests"))
	assert.NoError(err)
	// Init bm
	bm, err := NewBlobMaker(120*1024, filepath.Join(repoRoot, "prover/lib/compressor/compressor_dict.bin"))
	assert.NoError(err, "init should succeed")

	// Compress blocks
	cptBlock := 0
	for i, block := range testBlocks {
		// get a random from 1 to 5

		bSize := randIntn(5) + 1 // #nosec G404 -- false positive

		if cptBlock > bSize && i%3 == 0 {
			cptBlock = 0
			bm.StartNewBatch()
		}

		appended, err := bm.Write(block, false)
		if !appended || i == maxNbBlocks {
			assert.NoError(err, "append a valid block should not generate an error")
			break
		} else {
			cptBlock++
		}
	}
	return bm.Bytes()
}

func loadTestBlocks(testDataDir string) (testBlocks [][]byte, err error) {
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

func randIntn(n int) int {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return int(binary.BigEndian.Uint64(b[:]) % uint64(n))
}

func EmptyBlob(t require.TestingT) []byte {
	var headerB bytes.Buffer

	repoRoot, err := getRepoRootPath()
	assert.NoError(t, err)
	// Init bm
	bm, err := NewBlobMaker(1000, filepath.Join(repoRoot, "prover/lib/compressor/compressor_dict.bin"))
	assert.NoError(t, err)

	if _, err = bm.header.WriteTo(&headerB); err != nil {
		panic(err)
	}

	compressor, err := lzss.NewCompressor(LoadDict(t))
	assert.NoError(t, err)

	var bb bytes.Buffer
	if _, err = PackAlign(&bb, headerB.Bytes(), fr381.Bits-1, WithAdditionalInput(compressor.Bytes())); err != nil {
		panic(err)
	}
	return bb.Bytes()
}

func TinyTwoBatchBlob(t require.TestingT) []byte {
	repoRoot, err := getRepoRootPath()
	assert.NoError(t, err)
	testBlocks, err := loadTestBlocks(filepath.Join(repoRoot, "testdata/prover-v2/prover-execution/requests"))
	assert.NoError(t, err)
	// Init bm
	bm, err := NewBlobMaker(1000, filepath.Join(repoRoot, "prover/lib/compressor/compressor_dict.bin"))
	assert.NoError(t, err)

	for _, block := range testBlocks[:2] {
		ok, err := bm.Write(block, false)
		assert.NoError(t, err)
		assert.True(t, ok)
	}

	return bm.Bytes()
}
