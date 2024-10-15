package blob

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	fr381 "github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/linea-monorepo/prover/circuits/blobdecompression/v0/compress"
	v1 "github.com/consensys/linea-monorepo/prover/lib/compressor/blob/v1"
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

// DictionaryChecksum according to the given spec version
func DictionaryChecksum(dict []byte, version uint16) ([]byte, error) {
	switch version {
	case 1:
		return v1.MiMCChecksumPackedData(dict, 8)
	case 0:
		return compress.ChecksumPaddedBytes(dict, len(dict), hash.MIMC_BLS12_377.New(), fr381.Bits), nil
	}
	return nil, errors.New("unsupported version")
}

// GetRepoRootPath returns the root path of the repository
func GetRepoRootPath() (string, error) {
	// First, check if the deployment path exists
	deploymentPath := "/opt/linea"
	if _, err := os.Stat(deploymentPath); err == nil {
		return deploymentPath, nil
	}

	// If not found, fall back to checking the local development path based on the current directory
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	const repoName = "linea-monorepo"
	i := strings.LastIndex(wd, repoName)
	if i == -1 {
		return "", fmt.Errorf("could not find repo root. Current working directory: %s", wd)
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
