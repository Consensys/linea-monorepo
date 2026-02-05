package config

const (
	VerifyingKeyFileName      = "verifying_key.bin"
	CircuitFileName           = "circuit.bin"
	VerifierContractFileName  = "Verifier.sol"
	ManifestFileName          = "manifest.json"
	DefaultDictionaryFileName = "compressor_dict.bin"

	RequestsFromSubDir = "requests"
	RequestsToSubDir   = "responses"
	RequestsDoneSubDir = "requests-done"

	InProgressSuffix = "inprogress"
	FailSuffix       = "code"
	SuccessSuffix    = "success"

	// Extension to add in order to defer the job to the large prover
	LargeSuffix = "large"

	// Limitless prover stuffs
	BootstrapPartialSucessSuffix = "bootstrap.partial.success"
	WitnessDir                   = "/tmp/witness"
)

// Updated till osaka for now
var (
	ALL_MODULES = [16]string{
		"ARITH-OPS",
		"BLS_G1",
		"BLS_G2",
		"BLS_KZG",
		"BLS_PAIR",
		"ECDSA",
		"ECPAIRING",
		"ELLIPTIC_CURVES",
		"G2_CHECK",
		"HUB-KECCAK",
		"MODEXP_256",
		"MODEXP_LARGE",
		"P256",
		"SHA2",
		"STATIC",
		"TINY-STUFFS",
	}
)
