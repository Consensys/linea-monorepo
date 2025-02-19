package circuits

// CircuitID is a type to represent the different circuits.
// It is used to identify the circuit to be used in the prover.
type CircuitID string

const (
	ExecutionCircuitID                  CircuitID = "execution"
	ExecutionLargeCircuitID             CircuitID = "execution-large"
	BlobDecompressionV1CircuitID        CircuitID = "blob-decompression-v1"
	AggregationCircuitID                CircuitID = "aggregation"
	EmulationCircuitID                  CircuitID = "emulation"
	EmulationDummyCircuitID             CircuitID = "emulation-dummy"
	ExecutionDummyCircuitID             CircuitID = "execution-dummy"
	BlobDecompressionDummyCircuitID     CircuitID = "blob-decompression-dummy"
	PublicInputInterconnectionCircuitID CircuitID = "public-input-interconnection"
)

// MockCircuitID is a type to represent the different mock circuits.
type MockCircuitID int

const (
	MockCircuitIDExecution     MockCircuitID = 0
	MockCircuitIDDecompression MockCircuitID = 6789
	MockCircuitIDEmulation     MockCircuitID = 1
)
