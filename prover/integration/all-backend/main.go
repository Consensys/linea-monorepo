package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/test/unsafekzg"
	"github.com/consensys/linea-monorepo/prover/backend/aggregation"
	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/consensys/linea-monorepo/prover/lib/compressor/blob"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
)

var flagStart = flag.Int("start", 0, "starting block")
var flagEnd = flag.Int("end", 0, "ending block (inclusive)")

func main() {
	flag.Parse()

	var (
		t   test_utils.FakeTestingT
		req aggregation.Request
	)

	testPath, err := blob.GetRepoRootPath() // TODO @Tabaie move this function elsewhere
	require.NoError(t, err)
	testPath = filepath.Join(testPath, "prover", "integration", "all-backend")

	cfg := config.Config{
		Environment: "",
		Version:     "",
		LogLevel:    0,
		AssetsDir:   filepath.Join(testPath, "assets"),
		Controller:  config.Controller{},
		Execution: config.Execution{
			WithRequestDir: config.WithRequestDir{
				//RequestsRootDir: ,
			},
			ProverMode:         "dev",
			CanRunFullLarge:    false,
			ConflatedTracesDir: "",
		},
		BlobDecompression:          config.BlobDecompression{},
		Aggregation:                config.Aggregation{},
		PublicInputInterconnection: config.PublicInput{},
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

	ensureSRS(t, &cfg)

	// TODO @Tabaie remove
	*flagStart, *flagEnd = 37, 41

	// run execution proofs
	for i := *flagStart; i <= *flagEnd; i++ {
		var req execution.Request
		test_utils.LoadJson(
			t,
			findFile(t, filepath.Join(testPath, "prover-v2", "prover-execution", "requests"), fmt.Sprintf("%d-%d", i, i)),
			&req,
		)
		_, err := execution.Prove(&cfg, &req, false)
		require.NoError(t, err)
	}

	// find aggregation request file

	test_utils.LoadJson(
		t,
		findFile(t, filepath.Join(testPath, "prover-v2", "prover-aggregation", "requests"), fmt.Sprintf("%d-%d", *flagStart, *flagEnd)),
		&req,
	)

	_, err = aggregation.Prove(&cfg, &req)
	assert.NoError(t, err)
}

func findFile(t require.TestingT, dir, prefix string) string {

	logrus.Infof("searching for files with name prefix '%s' in folder '%s'", prefix, dir)
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			res := filepath.Join(dir, entry.Name())
			logrus.Infof("found '%s'", res)
			return res
		}
	}
	t.Errorf("unable to find file with name prefix \"%s\" in folder \"%s\"", prefix, dir)
	t.FailNow()
	return ""
}

func ensureSRS(t require.TestingT, cfg *config.Config) {
	/*srsStore, err := circuits.NewSRSStore(cfg.PathForSRS()); err == nil
	if _, err :=  // TODO check entries
		return
	}*/
	require.NoError(t, os.MkdirAll(cfg.PathForSRS(), 0600))
	var wg sync.WaitGroup
	createSRS := func(settings srsSpec) {
		canonical, lagrange, err := unsafekzg.NewSRS(&settings, unsafekzg.WithFSCache())
		require.NoError(t, err)
		f, err := os.OpenFile(filepath.Join(cfg.PathForSRS(), fmt.Sprintf("kzg_srs_canonical_%d_%s_aleo.memdump", settings.maxSize, settings.id.String())), os.O_WRONLY|os.O_CREATE, 0600) // not actually coming from Aleo
		require.NoError(t, err)
		require.NoError(t, canonical.WriteDump(f, settings.maxSize))
		f.Close()

		f, err = os.OpenFile(filepath.Join(cfg.PathForSRS(), fmt.Sprintf("kzg_srs_lagrange_%d_%s_aleo.memdump", settings.maxSize, settings.id.String())), os.O_WRONLY|os.O_CREATE, 0600) // not actually coming from Aleo
		require.NoError(t, err)
		require.NoError(t, lagrange.WriteDump(f, settings.maxSize))
		f.Close()

		wg.Done()
	}

	wg.Add(3)
	go createSRS(srsSpec{ecc.BLS12_377, 1 << 27})
	createSRS(srsSpec{ecc.BN254, 1 << 26})
	createSRS(srsSpec{ecc.BW6_761, 1 << 26})
	wg.Wait()
}
