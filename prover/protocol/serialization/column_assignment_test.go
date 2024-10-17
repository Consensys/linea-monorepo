package serialization

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/stretchr/testify/assert"
)

func TestSmartVectorCompression(t *testing.T) {

	svs := []smartvectors.SmartVector{
		smartvectors.NewConstant(field.Zero(), 16),
		smartvectors.ForTest(0, 0, 0, 0),
		smartvectors.NewConstant(field.NewElement(42), 16),
		smartvectors.ForTest(1, 2, 3, 4),
		smartvectors.LeftPadded(vector.ForTest(1, 2, 3, 4), field.Zero(), 16),
		smartvectors.LeftPadded(vector.ForTest(1, 2, 3, 4), field.One(), 16),
		smartvectors.RightPadded(vector.ForTest(1, 2, 3, 4), field.Zero(), 16),
		smartvectors.RightPadded(vector.ForTest(1, 2, 3, 4), field.One(), 16),
	}

	for i := range svs {
		t.Run(fmt.Sprintf("testcase-%v", i), func(t *testing.T) {
			t.Logf("original smartvector: %v\n", svs[i].Pretty())

			var (
				compressed   = CompressSmartVector(svs[i])
				decompressed = compressed.Decompress()
				recompressed = CompressSmartVector(decompressed)
			)

			assert.Equal(t, svs[i], decompressed)
			assert.Equal(t, compressed, recompressed)
		})
	}

}
