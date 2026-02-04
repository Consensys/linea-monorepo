package circuits

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeIsAllowedCircuitID(t *testing.T) {
	tests := []struct {
		name            string
		allowedCircuits []string
		expectedBitmask uint64
		expectError     bool
		errorContains   string
	}{
		{
			name: "mainnet configuration",
			allowedCircuits: []string{
				"execution",
				"execution-large",
				"execution-limitless",
				"data-availability-v2",
			},
			expectedBitmask: 120, // 0b01111000 = 2^3 + 2^4 + 2^5 + 2^6
			expectError:     false,
		},
		{
			name: "sepolia/testnet configuration",
			allowedCircuits: []string{
				"execution-dummy",
				"data-availability-dummy",
				"execution",
				"execution-large",
				"execution-limitless",
				"data-availability-v2",
			},
			expectedBitmask: 123, // 0b01111011 = 2^0 + 2^1 + 2^3 + 2^4 + 2^5 + 2^6
			expectError:     false,
		},
		{
			name: "devnet configuration",
			allowedCircuits: []string{
				"execution-dummy",
				"data-availability-dummy",
				"execution",
				"execution-large",
				"data-availability-v2",
			},
			expectedBitmask: 91, // 0b01011011 = 2^0 + 2^1 + 2^3 + 2^4 + 2^6
			expectError:     false,
		},
		{
			name: "integration-full configuration",
			allowedCircuits: []string{
				"execution-dummy",
				"data-availability-dummy",
				"execution",
			},
			expectedBitmask: 11, // 0b00001011 = 2^0 + 2^1 + 2^3
			expectError:     false,
		},
		{
			name: "integration-development configuration",
			allowedCircuits: []string{
				"execution-dummy",
				"data-availability-dummy",
				"data-availability-v2",
			},
			expectedBitmask: 67, // 0b01000011 = 2^0 + 2^1 + 2^6
			expectError:     false,
		},
		{
			name: "single circuit",
			allowedCircuits: []string{
				"execution",
			},
			expectedBitmask: 8, // 0b00001000 = 2^3
			expectError:     false,
		},
		{
			name:            "empty list",
			allowedCircuits: []string{},
			expectedBitmask: 0,
			expectError:     false,
		},
		{
			name: "unknown circuit name",
			allowedCircuits: []string{
				"execution",
				"nonexistent-circuit",
			},
			expectedBitmask: 0,
			expectError:     true,
			errorContains:   "unknown circuit name",
		},
		{
			name: "infrastructure circuit - aggregation",
			allowedCircuits: []string{
				"execution",
				"aggregation",
			},
			expectedBitmask: 0,
			expectError:     true,
			errorContains:   "infrastructure circuit",
		},
		{
			name: "infrastructure circuit - public-input-interconnection",
			allowedCircuits: []string{
				"execution",
				"public-input-interconnection",
			},
			expectedBitmask: 0,
			expectError:     true,
			errorContains:   "infrastructure circuit",
		},
		{
			name: "infrastructure circuit - emulation",
			allowedCircuits: []string{
				"execution",
				"emulation",
			},
			expectedBitmask: 0,
			expectError:     true,
			errorContains:   "infrastructure circuit",
		},
		{
			name: "all payload circuits (0-6)",
			allowedCircuits: []string{
				"execution-dummy",
				"data-availability-dummy",
				"emulation-dummy",
				"execution",
				"execution-large",
				"execution-limitless",
				"data-availability-v2",
			},
			expectedBitmask: 127, // 0b01111111 = all bits 0-6 set
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bitmask, err := ComputeIsAllowedCircuitID(tt.allowedCircuits)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Equal(t, uint64(0), bitmask)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBitmask, bitmask,
					"Expected bitmask %d (0b%b), got %d (0b%b)",
					tt.expectedBitmask, tt.expectedBitmask, bitmask, bitmask)
			}
		})
	}
}

func TestIsCircuitAllowed(t *testing.T) {
	tests := []struct {
		name      string
		bitmask   uint64
		circuitID uint
		expected  bool
	}{
		{
			name:      "mainnet allows execution",
			bitmask:   120, // mainnet
			circuitID: 3,   // execution
			expected:  true,
		},
		{
			name:      "mainnet disallows execution-dummy",
			bitmask:   120,
			circuitID: 0, // execution-dummy
			expected:  false,
		},
		{
			name:      "sepolia allows execution-dummy",
			bitmask:   123, // sepolia
			circuitID: 0,   // execution-dummy
			expected:  true,
		},
		{
			name:      "sepolia allows execution-large",
			bitmask:   123,
			circuitID: 4, // execution-large
			expected:  true,
		},
		{
			name:      "sepolia disallows emulation-dummy",
			bitmask:   123,
			circuitID: 2, // emulation-dummy
			expected:  false,
		},
		{
			name:      "zero bitmask disallows everything",
			bitmask:   0,
			circuitID: 3,
			expected:  false,
		},
		{
			name:      "all bits set allows everything",
			bitmask:   127, // 0b01111111
			circuitID: 6,   // data-availability-v2
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCircuitAllowed(tt.bitmask, tt.circuitID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAllowedCircuitNames(t *testing.T) {
	tests := []struct {
		name             string
		bitmask          uint64
		expectedCircuits []string
	}{
		{
			name:    "mainnet configuration",
			bitmask: 120,
			expectedCircuits: []string{
				"execution",
				"execution-large",
				"execution-limitless",
				"data-availability-v2",
			},
		},
		{
			name:    "sepolia configuration",
			bitmask: 123,
			expectedCircuits: []string{
				"execution-dummy",
				"data-availability-dummy",
				"execution",
				"execution-large",
				"execution-limitless",
				"data-availability-v2",
			},
		},
		{
			name:             "zero bitmask",
			bitmask:          0,
			expectedCircuits: []string{},
		},
		{
			name:    "only dummy circuits",
			bitmask: 3, // 0b00000011 = bits 0,1
			expectedCircuits: []string{
				"execution-dummy",
				"data-availability-dummy",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAllowedCircuitNames(tt.bitmask)

			// Sort both slices for comparison (map iteration order is random)
			assert.ElementsMatch(t, tt.expectedCircuits, result,
				"Expected circuits %v, got %v for bitmask %d (0b%b)",
				tt.expectedCircuits, result, tt.bitmask, tt.bitmask)
		})
	}
}

func TestRoundTripComputeAndCheck(t *testing.T) {
	// Test that computing a bitmask and then checking circuits gives expected results
	allowedCircuits := []string{
		"execution",
		"execution-large",
		"data-availability-v2",
	}

	bitmask, err := ComputeIsAllowedCircuitID(allowedCircuits)
	require.NoError(t, err)

	// These should be allowed
	assert.True(t, IsCircuitAllowed(bitmask, 3), "execution should be allowed")
	assert.True(t, IsCircuitAllowed(bitmask, 4), "execution-large should be allowed")
	assert.True(t, IsCircuitAllowed(bitmask, 6), "data-availability-v2 should be allowed")

	// These should NOT be allowed
	assert.False(t, IsCircuitAllowed(bitmask, 0), "execution-dummy should not be allowed")
	assert.False(t, IsCircuitAllowed(bitmask, 5), "execution-limitless should not be allowed")
	assert.False(t, IsCircuitAllowed(bitmask, 1), "data-availability-dummy should not be allowed")

	// Get allowed names and verify
	names := GetAllowedCircuitNames(bitmask)
	assert.ElementsMatch(t, allowedCircuits, names)
}

func TestGlobalCircuitIDMapping(t *testing.T) {
	// Verify the mapping contains expected entries
	assert.Equal(t, uint(0), GlobalCircuitIDMapping["execution-dummy"])
	assert.Equal(t, uint(1), GlobalCircuitIDMapping["data-availability-dummy"])
	assert.Equal(t, uint(2), GlobalCircuitIDMapping["emulation-dummy"])
	assert.Equal(t, uint(3), GlobalCircuitIDMapping["execution"])
	assert.Equal(t, uint(4), GlobalCircuitIDMapping["execution-large"])
	assert.Equal(t, uint(5), GlobalCircuitIDMapping["execution-limitless"])
	assert.Equal(t, uint(6), GlobalCircuitIDMapping["data-availability-v2"])
	assert.Equal(t, uint(8), GlobalCircuitIDMapping["emulation"])
	assert.Equal(t, uint(9), GlobalCircuitIDMapping["aggregation"])
	assert.Equal(t, uint(10), GlobalCircuitIDMapping["public-input-interconnection"])

	// Verify no duplicate IDs
	seen := make(map[uint]string)
	for name, id := range GlobalCircuitIDMapping {
		if existingName, exists := seen[id]; exists {
			t.Errorf("Duplicate circuit ID %d for circuits '%s' and '%s'", id, existingName, name)
		}
		seen[id] = name
	}

	// Verify we have exactly 10 circuits (IDs: 0,1,2,3,4,5,6,8,9,10)
	assert.Equal(t, 10, len(GlobalCircuitIDMapping))
}

// Example test showing how to use these functions for config validation
func TestExampleUsage(t *testing.T) {
	// Example: User wants to configure mainnet
	mainnetAllowedInputs := []string{
		"execution",
		"execution-large",
		"execution-limitless",
		"data-availability-v2",
	}

	// Compute the bitmask
	bitmask, err := ComputeIsAllowedCircuitID(mainnetAllowedInputs)
	require.NoError(t, err)
	assert.Equal(t, uint64(120), bitmask)

	// Verify it matches what's in the config
	t.Logf("Mainnet is_allowed_circuit_id = %d (binary: 0b%b)", bitmask, bitmask)

	// Check specific circuits
	assert.True(t, IsCircuitAllowed(bitmask, 3), "execution should be allowed in mainnet")
	assert.False(t, IsCircuitAllowed(bitmask, 0), "execution-dummy should not be allowed in mainnet")

	// Get human-readable list
	allowed := GetAllowedCircuitNames(bitmask)
	t.Logf("Mainnet allows: %v", allowed)
}
