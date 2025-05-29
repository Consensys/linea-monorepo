package serialization_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/serialization"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/consensys/linea-monorepo/prover/utils/test_utils"
	"github.com/consensys/linea-monorepo/prover/zkevm"
)

func TestZkEVM(t *testing.T) {

	z := test_utils.GetZkEVM()
	var b []byte
	var err error
	d := &zkevm.ZkEvm{}

	serTime := profiling.TimeIt(func() {
		b, err = serialization.Serialize(z)
		if err != nil {
			t.Fatal(err)
		}
	})

	desTime := profiling.TimeIt(func() {
		err := serialization.Deserialize(b, d)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Logf("serialization=%v deserialization=%v buffer-size=%v", serTime, desTime, len(b))
}
