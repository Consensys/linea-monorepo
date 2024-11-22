package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/test/unsafekzg"
	"github.com/consensys/linea-monorepo/prover/circuits"
	"github.com/consensys/linea-monorepo/prover/config"
	allbackend "github.com/consensys/linea-monorepo/prover/integration/all-backend"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/consensys/gnark/backend/plonk"
)

var flagLagrange = flag.String("lagrange", "", "12345_bls12377, 123_bn254")

func main() {

	var t test_utils.FakeTestingT

	flagConfig := "integration/all-backend/config-integration-light.toml"
	var flagLagrange []string
	argv := os.Args[1:]
	for len(argv) != 0 {
		switch argv[0] {
		case "-c", "-config":
			flagConfig = argv[1]
			argv = argv[2:]
		case "-l", "-lagrange":
			argv = argv[1:]
			n := 0
			for n < len(argv) && argv[n][0] != '-' {
				n++
			}
			flagLagrange = argv[:n]
			argv = argv[n:]
		default:
			logrus.Errorf("unknown parameter '%s'", argv[0])
			logrus.Errorf("usage ./makesrs")
			t.Errorf("unknown parameter '%s'", argv[0])
			t.FailNow()
		}
	}

	flag.Parse()

	allbackend.CdProver(t)

	cfg, err := config.NewConfigFromFile(flagConfig)
	require.NoError(t, err, "could not load config")

	store, err := circuits.NewSRSStore(cfg.PathForSRS())
	if err != nil {
		logrus.Errorf("creating SRS store: %v", err)
	}

	require.NoError(t, os.MkdirAll(cfg.PathForSRS(), 0600))
	var wg sync.WaitGroup
	createSRS := func(settings allbackend.SrsSpec) {
		logrus.Infof("checking for %s srs", settings.Id.String())
		canonicalSize, lagrangeSize := plonk.SRSSize(&settings)

		// check if it exists
		_, _, err = store.GetSRS(context.TODO(), &allbackend.SrsSpec{
			Id:      settings.Id,
			MaxSize: settings.MaxSize,
		})
		if err == nil {
			logrus.Infof("SRS found for %d-%s. Skipping creation", canonicalSize, settings.Id.String())
			return
		}
		logrus.Errorln(err)
		logrus.Infof("could not load %s SRS. Creating instead.", settings.Id.String())

		// it doesn't exist. create it.
		canonical, lagrange, err := unsafekzg.NewSRS(&settings /*, unsafekzg.WithFSCache()*/)
		require.NoError(t, err)
		f, err := os.OpenFile(filepath.Join(cfg.PathForSRS(), fmt.Sprintf("kzg_srs_canonical_%d_%s_aleo.memdump", canonicalSize, strings.Replace(settings.Id.String(), "_", "", 1))), os.O_WRONLY|os.O_CREATE, 0600) // not actually coming from Aleo
		require.NoError(t, err)
		require.NoError(t, canonical.WriteDump(f))
		f.Close()

		f, err = os.OpenFile(filepath.Join(cfg.PathForSRS(), fmt.Sprintf("kzg_srs_lagrange_%d_%s_aleo.memdump", lagrangeSize, strings.Replace(settings.Id.String(), "_", "", 1))), os.O_WRONLY|os.O_CREATE, 0600) // not actually coming from Aleo
		require.NoError(t, err)
		require.NoError(t, lagrange.WriteDump(f))
		f.Close()

		wg.Done()
	}

	if len(flagLagrange) == 0 {
		logrus.Info("no lagrange parameter. Setting up canonical SRS")

		wg.Add(3)
		go createSRS(allbackend.SrsSpec{ecc.BLS12_377, 1 << 27})
		createSRS(allbackend.SrsSpec{ecc.BN254, 1 << 26})
		createSRS(allbackend.SrsSpec{ecc.BW6_761, 1 << 26})
		wg.Wait()
		return
	}

	logrus.Info("Lagrange parameter set. Assuming canonical SRS is set.")

	nameToCurve := map[string]ecc.ID{
		"bls12377": ecc.BLS12_377,
		"bn254":    ecc.BN254,
		"bw6761":   ecc.BW6_761,
	}

	writeLagrange := func(param string) {
		split := strings.Split(param, "_")
		if len(split) != 2 {
			panic("bad format")
		}
		size, err := strconv.Atoi(split[0])
		require.NoError(t, err)

		var match ecc.ID
		for k, v := range nameToCurve {
			if strings.HasPrefix(k, split[1]) { // a match
				logrus.Infof("argument '%s' matches curve name '%s'", split[1], k)
				if match != ecc.UNKNOWN {
					t.Errorf("multiple matches")
					t.FailNow()
				}
				match = v
			}
		}
		require.NotEqual(t, ecc.UNKNOWN, match, "argument '%s' doesn't match any supported curves", split[1])

		_, lagrange, err := store.GetSRS(context.TODO(), &allbackend.SrsSpec{match, size})
		require.NoError(t, err)
		f, err := os.OpenFile(filepath.Join(cfg.PathForSRS(), fmt.Sprintf("kzg_srs_lagrange_%d_%s_aleo.memdump", size, strings.ReplaceAll(match.String(), "_", ""))), os.O_WRONLY|os.O_CREATE, 0600) // not actually coming from Aleo
		require.NoError(t, err)
		require.NoError(t, lagrange.WriteDump(f, size))
		f.Close()
		wg.Done()
	}

	wg.Add(len(flagLagrange))
	for _, lagrange := range flagLagrange[1:] {
		go writeLagrange(lagrange)
	}
	writeLagrange(flagLagrange[0])
	wg.Wait()
}
