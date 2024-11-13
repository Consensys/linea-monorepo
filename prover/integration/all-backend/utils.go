package allbackend

import (
	"os"
	"path/filepath"

	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func CdProver(t require.TestingT) {
	rootDir, err := blob.GetRepoRootPath()
	require.NoError(t, err)
	rootDir = filepath.Join(rootDir, "prover")
	logrus.Infof("switching working directory to '%s'", rootDir)
	require.NoError(t, os.Chdir(rootDir))

}
