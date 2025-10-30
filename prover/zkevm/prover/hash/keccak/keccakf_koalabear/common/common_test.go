package common_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/zkevm/prover/hash/keccak/keccakf_koalabear/common"
)

func TestDecomposeUint32(t *testing.T) {
	const base = 14641 //11^4
	const nb = 2
	n := uint64(15972)            // 11^4+11^3
	expected := []uint64{1331, 1} // precomputed expected result
	result := common.DecomposeU64(n, base, nb)
	if len(result) != nb {
		t.Fatalf("expected %v limbs, but got %v", nb, len(result))
	}
	for i := 0; i < nb; i++ {
		if result[i].Uint64() != uint64(expected[i]) {
			t.Errorf("limb %v: expected %v, but got %v", i, expected[i], result[i].Uint64())
		}
	}
}

func TestCleanBaseChi(t *testing.T) {
	const base = 11
	n := []uint64{22, 33, 44}     // in base dirty BaseChi
	expected := []uint64{1, 0, 0} // precomputed expected result after cleaning
	var nField []field.Element    // convert to field elements
	for _, v := range n {
		nField = append(nField, field.NewElement(v))
	}
	cleaned := common.CleanBaseChi(nField)

	for i := 0; i < len(expected); i++ {
		if cleaned[i].Uint64() != expected[i] {
			t.Errorf("element %v: expected %v, but got %v", i, expected[i], cleaned[i].Uint64())
		}
	}
}
