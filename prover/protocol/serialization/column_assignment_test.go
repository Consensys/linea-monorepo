package serialization

import (
	"fmt"

	"os"
	"strconv"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
	"github.com/stretchr/testify/assert"
)

// Global svs array to be reused in both tests
var svs = []smartvectors.SmartVector{
	smartvectors.NewConstant(field.Zero(), 16),
	smartvectors.ForTest(0, 0, 0, 0),
	smartvectors.NewConstant(field.NewElement(42), 16),
	smartvectors.ForTest(1, 2, 3, 4),
	smartvectors.LeftPadded(vector.ForTest(1, 2, 3, 4), field.Zero(), 16),
	smartvectors.LeftPadded(vector.ForTest(1, 2, 3, 4), field.One(), 16),
	smartvectors.RightPadded(vector.ForTest(1, 2, 3, 4), field.Zero(), 16),
	smartvectors.RightPadded(vector.ForTest(1, 2, 3, 4), field.One(), 16),
}

func TestSerializeAndDeserializeAssignment(t *testing.T) {
	// Create a sample WAssignment using svs
	wAssignment := collection.NewMapping[ifaces.ColID, smartvectors.SmartVector]()
	for i, sv := range svs {
		wAssignment.InsertNew(ifaces.ColID(fmt.Sprintf("col%d", i+1)), sv)
	}

	// Serialize the WAssignment
	numChunks := 4
	serializedChunks := SerializeAssignment(wAssignment, numChunks)

	// Compress the serialized chunks
	compressedChunks := CompressChunks(serializedChunks)

	// Write compressed chunks to temporary files
	tempDir, err := os.MkdirTemp("", "serialization_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	for i, chunk := range compressedChunks {
		chunkPath := tempDir + "/chunk_" + strconv.Itoa(i)
		err := os.WriteFile(chunkPath, chunk, 0600) // Use 0600 permissions
		assert.NoError(t, err)
	}

	// Deserialize the WAssignment from the temporary files
	deserializedAssignment, err := DeserializeAssignment(tempDir+"/chunk_", numChunks)
	assert.NoError(t, err)

	// Verify that the deserialized WAssignment matches the original WAssignment
	assert.Equal(t, hashWAssignment(wAssignment), hashWAssignment(deserializedAssignment))
}

func TestSmartVectorCompression(t *testing.T) {
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
