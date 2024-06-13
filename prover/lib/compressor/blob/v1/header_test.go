package v1

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHeaderByteSize(t *testing.T) {
	var bb bytes.Buffer
	const maxSize = 200
	var _batchSizes, _currBatchBlocksLen [maxSize]int
	for i := 0; i < 300; i++ {
		batchSizes := _batchSizes[:randIntn(maxSize)]
		currBatchBlocksLen := _currBatchBlocksLen[:randIntn(maxSize)]

		for j := range batchSizes {
			batchSizes[j] = randIntn(0x100000)
		}
		for j := range currBatchBlocksLen {
			currBatchBlocksLen[j] = randIntn(0x10000)
		}

		header := Header{
			BatchSizes:         batchSizes,
			currBatchBlocksLen: currBatchBlocksLen,
		}

		bb.Reset()
		expectedSize := header.ByteSize()
		_, err := header.WriteTo(&bb)
		assert.NoError(t, err)
		assert.Equal(t, expectedSize, bb.Len())
	}
}
