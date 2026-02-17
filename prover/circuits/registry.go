package circuits

import "fmt"

// CircuitID is a type to represent the different circuits.
// It is used to identify the circuit to be used in the prover.
type CircuitID string

const (
	ExecutionCircuitID          CircuitID = "execution"
	ExecutionLargeCircuitID     CircuitID = "execution-large"
	ExecutionLimitlessCircuitID CircuitID = "execution-limitless"
	ExecutionDummyCircuitID     CircuitID = "execution-dummy"

	DataAvailabilityV2CircuitID    CircuitID = "data-availability-v2"
	DataAvailabilityDummyCircuitID CircuitID = "data-availability-dummy"

	AggregationCircuitID CircuitID = "aggregation"

	EmulationCircuitID      CircuitID = "emulation"
	EmulationDummyCircuitID CircuitID = "emulation-dummy"

	PublicInputInterconnectionCircuitID CircuitID = "public-input-interconnection"

	InvalidityNonceBalanceCircuitID         CircuitID = "invalidity-nonce-balance"
	InvalidityPrecompileLogsCircuitID       CircuitID = "invalidity-precompile-logs"
	InvalidityFilteredAddressCircuitID      CircuitID = "invalidity-filtered-address"
	InvalidityNonceBalanceDummyCircuitID    CircuitID = "invalidity-nonce-balance-dummy"
	InvalidityPrecompileLogsDummyCircuitID  CircuitID = "invalidity-precompile-logs-dummy"
	InvalidityFilteredAddressDummyCircuitID CircuitID = "invalidity-filtered-address-dummy"
)

// MockCircuitID is a type to represent the different mock circuits.
type MockCircuitID int

const (
	MockCircuitIDExecution                 MockCircuitID = 0
	MockCircuitIDDecompression             MockCircuitID = 6789
	MockCircuitIDEmulation                 MockCircuitID = 1
	MockCircuitIDInvalidityNonceBalance    MockCircuitID = 2
	MockCircuitIDInvalidityPrecompileLogs  MockCircuitID = 3
	MockCircuitIDInvalidityFilteredAddress MockCircuitID = 4
)

// GlobalCircuitIDMapping defines the fixed mapping of circuit names to circuit IDs.
// This order is canonical and must remain stable across versions.
//
// RELATIONSHIP TO is_allowed_circuit_id:
//
// The is_allowed_circuit_id config field is a bitmask that controls which INNER PAYLOAD
// circuits (execution, decompression, and invalidity variants) can be aggregated in a
// given environment.
//
//   - Bit i (LSb to MSb) indicates whether circuit ID i is allowed
//   - Only circuits 0-12 are used in the bitmask (inner payload circuits)
//   - Circuits 14-16 (emulation, aggregation, PI-interconnection) are infrastructure
//     circuits and should NOT be included in the bitmask
//
// HOW TO COMPUTE is_allowed_circuit_id:
//
// For each circuit you want to allow, set the corresponding bit:
//
//	is_allowed_circuit_id = Σ(2^circuit_id) for each allowed circuit
//
// EXAMPLES:
//
// Mainnet (production circuits only, with invalidity):
//
//	execution-dummy (ID 0, bit 0)                   = 0 → DISALLOWED
//	data-availability-dummy (ID 1, bit 1)            = 0 → DISALLOWED
//	emulation-dummy (ID 2, bit 2)                    = 0 → DISALLOWED
//	execution (ID 3, bit 3)                          = 1 → ALLOWED
//	execution-large (ID 4, bit 4)                    = 1 → ALLOWED
//	execution-limitless (ID 5, bit 5)                = 1 → ALLOWED
//	data-availability-v2 (ID 6, bit 6)               = 1 → ALLOWED
//	invalidity-nonce-balance-dummy (ID 7, bit 7)     = 0 → DISALLOWED
//	invalidity-precompile-logs-dummy (ID 8, bit 8)   = 0 → DISALLOWED
//	invalidity-filtered-address-dummy (ID 9, bit 9)  = 0 → DISALLOWED
//	invalidity-nonce-balance (ID 10, bit 10)         = 1 → ALLOWED
//	invalidity-precompile-logs (ID 11, bit 11)       = 1 → ALLOWED
//	invalidity-filtered-address (ID 12, bit 12)      = 1 → ALLOWED
//	Binary: 0b1110001111000 = 7288 (decimal)
//	is_allowed_circuit_id = 7288
//
// Use ComputeIsAllowedCircuitID() to calculate the bitmask from circuit names.
var GlobalCircuitIDMapping = map[string]uint{
	// Dummy circuits (bits 0-2) - for testing environments
	"execution-dummy":         0,
	"data-availability-dummy": 1,
	"emulation-dummy":         2,

	// Production payload circuits (bits 3-6) - aggregated inner proofs
	"execution":            3,
	"execution-large":      4,
	"execution-limitless":  5,
	"data-availability-v2": 6,

	// Invalidity dummy circuits (bits 7-9) - for testing environments
	"invalidity-nonce-balance-dummy":    7,
	"invalidity-precompile-logs-dummy":  8,
	"invalidity-filtered-address-dummy": 9,

	// Invalidity production circuits (bits 10-12) - aggregated invalidity proofs
	"invalidity-nonce-balance":    10,
	"invalidity-precompile-logs":  11,
	"invalidity-filtered-address": 12,

	// Infrastructure circuits (bits 14-16) - NOT included in is_allowed_circuit_id bitmask
	"emulation":                    14,
	"aggregation":                  15,
	"public-input-interconnection": 16,
}

// GetAllCircuitNames returns all circuit names in the global mapping, sorted by circuit ID.
func GetAllCircuitNames() []string {
	// Create reverse mapping
	idToName := make(map[uint]string)
	maxID := uint(0)
	for name, id := range GlobalCircuitIDMapping {
		idToName[id] = name
		if id > maxID {
			maxID = id
		}
	}

	// Build sorted list
	result := make([]string, 0, len(GlobalCircuitIDMapping))
	for i := uint(0); i <= maxID; i++ {
		if name, exists := idToName[i]; exists {
			result = append(result, name)
		}
	}
	return result
}

// ComputeIsAllowedCircuitID computes the is_allowed_circuit_id bitmask from a list of
// allowed circuit names. This is useful for:
// - Validating config files
// - Generating new configs for different environments
// - Testing different circuit combinations
//
// Example:
//
//	allowed := []string{"execution", "execution-large", "data-availability-v2"}
//	bitmask := circuits.ComputeIsAllowedCircuitID(allowed)
//	// bitmask = 88 (binary: 0b01011000, bits 3,4,6 set)
//
// Returns an error if any circuit name is not found in GlobalCircuitIDMapping.
func ComputeIsAllowedCircuitID(allowedCircuits []string) (uint64, error) {
	var bitmask uint64 = 0

	for _, name := range allowedCircuits {
		id, exists := GlobalCircuitIDMapping[name]
		if !exists {
			return 0, fmt.Errorf("unknown circuit name: %s", name)
		}

		// Infrastructure circuits (14+) should not be in the bitmask
		if id >= 14 {
			return 0, fmt.Errorf("circuit '%s' (ID %d) is an infrastructure circuit and should not be included in is_allowed_circuit_id", name, id)
		}

		bitmask |= (1 << id)
	}

	return bitmask, nil
}

// IsCircuitAllowed checks if a circuit ID is allowed according to the given bitmask.
// This is the reverse operation of ComputeIsAllowedCircuitID.
//
// Example:
//
//	bitmask := uint64(120) // 0b01111000 - mainnet configuration
//	allowed := circuits.IsCircuitAllowed(bitmask, 3) // execution
//	// allowed = true
func IsCircuitAllowed(bitmask uint64, circuitID uint) bool {
	return (bitmask & (1 << circuitID)) != 0
}

// GetAllowedCircuitNames returns a list of circuit names that are allowed according
// to the given bitmask. This is useful for debugging and displaying which circuits
// are enabled in a configuration.
//
// Example:
//
//	bitmask := uint64(120) // mainnet
//	names := circuits.GetAllowedCircuitNames(bitmask)
//	// names = ["execution", "execution-large", "execution-limitless", "data-availability-v2"]
func GetAllowedCircuitNames(bitmask uint64) []string {
	var allowed []string

	for name, id := range GlobalCircuitIDMapping {
		// Only check payload circuits (0-12)
		if id < 14 && IsCircuitAllowed(bitmask, id) {
			allowed = append(allowed, name)
		}
	}

	return allowed
}
