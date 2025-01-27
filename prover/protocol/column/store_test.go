package column_test

import (
	"fmt"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/stretchr/testify/assert"
)

func TestStore(t *testing.T) {

	store := column.NewStore()

	assert.False(t, store.Exists("a"))
	assert.Panics(t, func() {
		store.IsIgnored("a")
	})

	store.AddToRound(0, "a", 4, column.Committed)

	assert.Len(t, store.AllKeys(), 1)
	assert.True(t, store.Exists("a"))
	assert.False(t, store.IsIgnored("a"))

	store.MarkAsIgnored("a")

	assert.True(t, store.IsIgnored("a"))

}

func TestNextPowerOfTwo(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{1, 1},
		{2, 2},
		{5, 8},
		{12, 16},
		{20, 32},
		{33, 64},
		{100, 128},
		{255, 256},
		{500, 512},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("NextPowerOfTwo(%d)", test.input), func(t *testing.T) {
			result := utils.NextPowerOfTwo(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}
