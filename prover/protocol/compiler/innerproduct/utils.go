package innerproduct

import (
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizardutils"
)

const innerProdStr = "INNERPRODUCT"

func deriveName[R ~string](ss ...any) R {
	ss = append([]any{innerProdStr}, ss...)
	return wizardutils.DeriveName[R](ss...)
}
