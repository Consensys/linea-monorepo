package limitless

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/consensys/linea-monorepo/prover/backend/execution"
	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/consensys/linea-monorepo/prover/config"
)

func Test_Bootstrap(t *testing.T) {
	var (
		reqFile      = files.MustRead("/home/ubuntu/testing-sepolia-0.8.0-rc8.1/prover-execution/requests/16303865-16303866-etv0.2.0-stv2.2.2-getZkProof.json")
		cfgFilePath  = "/home/ubuntu/linea-monorepo/prover/config/config-test.toml"
		req          = &execution.Request{}
		reqDecodeErr = json.NewDecoder(reqFile).Decode(req)
		cfg, cfgErr  = config.NewConfigFromFileUnchecked(cfgFilePath)
	)

	if reqDecodeErr != nil {
		t.Fatalf("could not read the request file: %v", reqDecodeErr)
	}

	if cfgErr != nil {
		t.Fatalf("could not read the config file: err=%v, cfg=%++v", cfgErr, cfg)
	}

	t.Logf("loaded config: %++v", cfg)

	t.Logf("[%v] running the bootstrapper\n", time.Now())

	_, _, err := InitBootstrapper(cfg, req, 1<<28)
	if err != nil {
		t.Fatal("Failed running the bootstrapper")
	}
}
