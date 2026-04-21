package utils_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover-ray/utils"
)

func TestHexEncodeToString(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{"Empty slice", []byte{}, "0x"},
		{"Single byte", []byte{0xab}, "0xab"},
		{"Multiple bytes", []byte{0xab, 0xcd, 0xef}, "0xabcdef"},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.HexEncodeToString(tt.input)
			if got != tt.want {
				t.Errorf("test case %d: want %s, got %s", i, tt.want, got)
			}
		})
	}
}
