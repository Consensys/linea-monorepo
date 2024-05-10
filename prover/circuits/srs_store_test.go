package circuits

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSRSStore(t *testing.T) {
	assert := require.New(t)

	srsStore, err := NewSRSStore("../prover-assets/kzgsrs")
	assert.NoError(err)

	assert.True(len(srsStore.entries) > 0)

	// log the entries
	for curveID, entries := range srsStore.entries {
		t.Logf("curveID %s\n", curveID)
		for _, entry := range entries {
			t.Logf("entry %v\n", entry)
		}
	}
}
