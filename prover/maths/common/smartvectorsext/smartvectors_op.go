package smartvectorsext

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors/vectorext"
)

// ForTest returns a witness from a explicit litteral assignement
func ForTestExt(xs ...int) smartvectors.SmartVector {
	return NewRegularExt(vectorext.ForTest(xs...))
}
