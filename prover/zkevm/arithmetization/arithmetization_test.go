package arithmetization

import (
	"testing"

	"github.com/consensys/go-corset/pkg/util/collection/typed"
)

func TestSanityCheck(t *testing.T) {
	// Test case 1: successful sanity check
	metadata := typed.NewMap(map[string]any{
		"Metadata": map[string]any{
			"chainId": 59144,
			"lineCounts": map[string]any{
				"BLOCK_L2_L1_LOGS": 10,
			},
		},
	})
	opts := &SanityCheckOptions{
		ChainID:                59144,
		NbAllL2L1MessageHashes: 10,
	}
	sanityCheck(metadata, opts)
	// If no panic, test passed

	// Test case 2: mismatched chainID
	metadata = typed.NewMap(map[string]any{
		"Metadata": map[string]any{
			"chainId": 59145,
			"lineCounts": map[string]any{
				"BLOCK_L2_L1_LOGS": 10,
			},
		},
	})
	opts = &SanityCheckOptions{
		ChainID:                59144,
		NbAllL2L1MessageHashes: 10,
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for mismatched chainID")
		}
	}()
	sanityCheck(metadata, opts)

	// Test case 3: mismatched BLOCK_L2_L1_LOGS
	metadata = typed.NewMap(map[string]any{
		"Metadata": map[string]any{
			"chainId": 59144,
			"lineCounts": map[string]any{
				"BLOCK_L2_L1_LOGS": 20,
			},
		},
	})
	opts = &SanityCheckOptions{
		ChainID:                59144,
		NbAllL2L1MessageHashes: 10,
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for mismatched BLOCK_L2_L1_LOGS")
		}
	}()
	sanityCheck(metadata, opts)

	// Test case 4: missing chainId
	metadata = typed.NewMap(map[string]any{
		"Metadata": map[string]any{
			"lineCounts": map[string]any{
				"BLOCK_L2_L1_LOGS": 10,
			},
		},
	})
	opts = &SanityCheckOptions{
		ChainID:                59144,
		NbAllL2L1MessageHashes: 10,
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for missing chainId")
		}
	}()
	sanityCheck(metadata, opts)

	// Test case 5: missing BLOCK_L2_L1_LOGS
	metadata = typed.NewMap(map[string]any{
		"Metadata": map[string]any{
			"chainId":    59144,
			"lineCounts": map[string]any{},
		},
	})
	opts = &SanityCheckOptions{
		ChainID:                59144,
		NbAllL2L1MessageHashes: 10,
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for missing BLOCK_L2_L1_LOGS")
		}
	}()
	sanityCheck(metadata, opts)
}
