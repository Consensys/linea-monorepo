package parallel_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/utils/parallel"
)

func TestParChunky(t *testing.T) {

	testArr := make([]int, 4096)

	worker := func(start, stop int) {
		for i := start; i < stop; i++ {
			testArr[i] += 1
		}
	}

	parallel.ExecuteChunky(4096, worker)

	for i := range testArr {
		if testArr[i] != 1 {
			t.Logf("went %v times over the pos %v", testArr[i], i)
			t.Fail()
		}
	}

}
