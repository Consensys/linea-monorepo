package common_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf_koalabear/common"
)

func TestDecomposeUint32(t *testing.T) {
	const base = 14641 //11^4
	const nb = 2
	n := uint64(15972)            // 11^4+11^3
	expected := []uint64{1331, 1} // precomputed expected result
	result := common.DecomposeUint32(n, base, nb)
	if len(result) != nb {
		t.Fatalf("expected %v limbs, but got %v", nb, len(result))
	}
	for i := 0; i < nb; i++ {
		if result[i].Uint64() != uint64(expected[i]) {
			t.Errorf("limb %v: expected %v, but got %v", i, expected[i], result[i].Uint64())
		}
	}
}
