package main

import (
	"io/fs"
	"path/filepath"

	"github.com/consensys/accelerated-crypto-monorepo/backend/config"
	"github.com/consensys/accelerated-crypto-monorepo/backend/files"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/compiler/dummy"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils/profiling"
	"github.com/sirupsen/logrus"

	"github.com/consensys/accelerated-crypto-monorepo/zkevm"
	"github.com/consensys/accelerated-crypto-monorepo/zkevm/define"
)

const TEST_CASE_DIR = "../../testdata/traces"

func main() {

	profiling.SKIP_PROFILING = true

	config.InitLogger()

	filepath.Walk(TEST_CASE_DIR, func(inp string, info fs.FileInfo, _ error) error {

		if info == nil || info.IsDir() {
			// Ignore calls for the directory
			return nil
		}

		f := files.MustReadCompressed(inp)

		f.Close()

		compiled := wizard.Compile(
			zkevm.WrapDefine(define.ZkEVMDefine),
			dummy.Compile,
		)

		proof := wizard.Prove(compiled, func(run *wizard.ProverRuntime) { zkevm.AssignFromCorset(f, run) })
		err := wizard.Verify(compiled, proof)
		if err != nil {
			logrus.Errorf("error for file: %s,  err: %s", inp, err)
		}

		// nil to indicate that WalkDir can continue
		return nil
	})

}
