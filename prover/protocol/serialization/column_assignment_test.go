package serialization

import (
	"fmt"
	"os"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/utils/collection"
	"github.com/stretchr/testify/assert"
)

// Global svs array to be reused in tests
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
	serializedChunks, err := SerializeAssignment(wAssignment, numChunks)
	assert.NoError(t, err, "SerializeAssignment failed")

	// Compress the serialized chunks
	compressedChunks, err := CompressChunks(serializedChunks)
	assert.NoError(t, err, "CompressChunks failed")

	// Write compressed chunks to temporary files
	tempDir, err := os.MkdirTemp("", "serialization_test")
	assert.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempDir)

	for i, chunk := range compressedChunks {
		chunkPath := fmt.Sprintf("%s/chunk_%d", tempDir, i)
		err := os.WriteFile(chunkPath, chunk, 0600)
		assert.NoError(t, err, "Failed to write chunk %d", i)
	}

	// Deserialize the WAssignment from the temporary files
	deserializedAssignment, err := DeserializeAssignment(tempDir+"/chunk_", numChunks)
	assert.NoError(t, err, "DeserializeAssignment failed")

	// Verify that the deserialized WAssignment matches the original
	assert.Equal(t, hashWAssignment(wAssignment), hashWAssignment(deserializedAssignment), "Hash mismatch after deserialization")
}

func TestWAssignmentSerialization(t *testing.T) {
	// Test non-empty WAssignment
	w := collection.NewMapping[ifaces.ColID, smartvectors.SmartVector]()
	for i, sv := range svs[:2] {
		w.InsertNew(ifaces.ColID(fmt.Sprintf("col%d", i+1)), sv)
	}

	data, err := serializeWAssignment(w)
	assert.NoError(t, err, "serializeWAssignment failed")
	assert.NotEmpty(t, data, "Serialized data empty")

	newW, err := deserializeWAssignment(data)
	assert.NoError(t, err, "deserializeWAssignment failed")
	assert.Equal(t, w, newW, "Round-trip mismatch")

	// Test empty WAssignment
	emptyW := collection.NewMapping[ifaces.ColID, smartvectors.SmartVector]()
	data, err = serializeWAssignment(emptyW)
	assert.NoError(t, err, "serializeWAssignment failed for empty WAssignment")
	newEmptyW, err := deserializeWAssignment(data)
	assert.NoError(t, err, "deserializeWAssignment failed for empty WAssignment")
	assert.Equal(t, 0, newEmptyW.Len(), "Empty WAssignment should have zero entries")
}

func TestSmartVectorCompression(t *testing.T) {
	for i, sv := range svs {
		t.Run(fmt.Sprintf("testcase-%v", i), func(t *testing.T) {
			t.Logf("original smartvector: %v\n", sv.Pretty())

			compressed := CompressSmartVector(sv)
			data, err := compressed.Serialize()
			assert.NoError(t, err, "Serialize CompressedSmartVector failed")

			var newCompressed CompressedSmartVector
			err = newCompressed.Deserialize(data)
			assert.NoError(t, err, "Deserialize CompressedSmartVector failed")

			decompressed := newCompressed.Decompress()
			recompressed := CompressSmartVector(decompressed)

			assert.Equal(t, sv, decompressed, "Decompressed SmartVector mismatch")
			assert.Equal(t, compressed, recompressed, "Recompressed SmartVector mismatch")
		})
	}
}

func TestSerializeAssignmentErrors(t *testing.T) {
	// Test invalid numChunks
	w := collection.NewMapping[ifaces.ColID, smartvectors.SmartVector]()
	_, err := SerializeAssignment(w, 0)
	assert.Error(t, err, "Should fail with invalid numChunks")
	assert.Contains(t, err.Error(), "invalid numChunks")

	// Test oversized WAssignment
	largeW := collection.NewMapping[ifaces.ColID, smartvectors.SmartVector]()
	for i := 0; i < MaxMapPairs+1; i++ {
		largeW.InsertNew(ifaces.ColID(fmt.Sprintf("col%d", i)), smartvectors.NewConstant(field.NewElement(0), 1))
	}
	_, err = SerializeAssignment(largeW, 1)
	assert.Error(t, err, "Should reject oversized map")
	assert.Contains(t, err.Error(), "map size")
}

func TestDeserializeAssignmentErrors(t *testing.T) {
	// Create a temporary directory with no files
	tempDir, err := os.MkdirTemp("", "serialization_test")
	assert.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempDir)

	// Test deserialization with missing files
	_, err = DeserializeAssignment(tempDir+"/chunk_", 1)
	assert.Error(t, err, "Should fail with missing chunk file")
	assert.Contains(t, err.Error(), "failed to read chunk")

	// Test invalid chunk data
	chunkPath := fmt.Sprintf("%s/chunk_0", tempDir)
	err = os.WriteFile(chunkPath, []byte("invalid data"), 0600)
	assert.NoError(t, err, "Failed to write invalid chunk")
	_, err = DeserializeAssignment(tempDir+"/chunk_", 1)
	assert.Error(t, err, "Should fail with invalid chunk data")
}
