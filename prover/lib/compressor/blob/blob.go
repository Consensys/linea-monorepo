package blob

import (
	"bytes"
	"errors"
	"github.com/ethereum/go-ethereum/core/types"
	"os"
	"path/filepath"
	"strings"

	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/dictionary"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob/encode"
	v0 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v0"
	v1 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1"
	"github.com/ethereum/go-ethereum/rlp"
)

func GetVersion(blob []byte) uint16 {
	if len(blob) < 3 {
		return 0
	}

	if blob[0] == 0x3f && blob[1] == 0xff && blob[2]&0xc0 == 0xc0 {
		return 1
	}
	return 0
}

// GetRepoRootPath assumes that current working directory is within the repo
func GetRepoRootPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	const repoName = "linea-monorepo"
	i := strings.LastIndex(wd, repoName)
	if i == -1 {
		return "", errors.New("could not find repo root")
	}
	i += len(repoName)
	return wd[:i], nil
}

func GetDict() ([]byte, error) {
	repoRoot, err := GetRepoRootPath()
	if err != nil {
		return nil, err
	}
	dictPath := filepath.Join(repoRoot, "prover/lib/compressor/compressor_dict.bin")
	return os.ReadFile(dictPath)
}

func DecompressBlob(blob []byte, dictStore dictionary.Store) ([]byte, error) {
	vsn := GetVersion(blob)
	var (
		blockDecoder func(*bytes.Reader) (encode.DecodedBlockData, error)
		blocks       [][]byte
		err          error
	)
	switch vsn {
	case 0:
		_, _, blocks, err = v0.DecompressBlob(blob, dictStore)
		blockDecoder = v0.DecodeBlockFromUncompressed
	case 1:
		_, _, blocks, err = v1.DecompressBlob(blob, dictStore)
		blockDecoder = v1.DecodeBlockFromUncompressed
	default:
		return nil, errors.New("unrecognized blob version")
	}

	if err != nil {
		return nil, err
	}
	blockObjs := make([]*types.Block, len(blocks))
	var decodedBlock encode.DecodedBlockData
	for i, block := range blocks {
		if decodedBlock, err = blockDecoder(bytes.NewReader(block)); err != nil {
			return nil, err
		}
		blockObjs[i] = decodedBlock.ToStd()
	}
	return rlp.EncodeToBytes(blockObjs)
}
