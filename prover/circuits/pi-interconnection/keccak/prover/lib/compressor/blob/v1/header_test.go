package v1_test

import (
	"bytes"
	"testing"

	v1 "github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/lib/compressor/blob/v1"
	"github.com/consensys/linea-monorepo/prover/circuits/pi-interconnection/keccak/prover/lib/compressor/blob/v1/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestHeaderByteSize(t *testing.T) {
	var bb bytes.Buffer
	const maxSize = 200
	var _batchSizes, _currBatchBlocksLen [maxSize]int
	for i := 0; i < 300; i++ {
		batchSizes := _batchSizes[:test_utils.RandIntn(maxSize)]
		currBatchBlocksLen := _currBatchBlocksLen[:test_utils.RandIntn(maxSize)]

		for j := range batchSizes {
			batchSizes[j] = test_utils.RandIntn(0x100000)
		}
		for j := range currBatchBlocksLen {
			currBatchBlocksLen[j] = test_utils.RandIntn(0x10000)
		}

		header := v1.Header{
			BatchSizes:         batchSizes,
			CurrBatchBlocksLen: currBatchBlocksLen,
		}

		bb.Reset()
		expectedSize := header.ByteSize()
		_, err := header.WriteTo(&bb)
		assert.NoError(t, err)
		assert.Equal(t, expectedSize, bb.Len())
	}
}
