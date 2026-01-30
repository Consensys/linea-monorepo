package circuits

// CircuitID is a type to represent the different circuits.
// It is used to identify the circuit to be used in the prover.
type CircuitID string

const (
	ExecutionCircuitID          CircuitID = "execution"
	ExecutionLargeCircuitID     CircuitID = "execution-large"
	ExecutionLimitlessCircuitID CircuitID = "execution-limitless"
	ExecutionDummyCircuitID     CircuitID = "execution-dummy"

	BlobDecompressionV0CircuitID    CircuitID = "blob-decompression-v0"
	BlobDecompressionV1CircuitID    CircuitID = "blob-decompression-v1"
	BlobDecompressionDummyCircuitID CircuitID = "blob-decompression-dummy"

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
