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

	InProgressSufix = "inprogress"
	FailSuffix      = "code"
	SuccessSuffix   = "success"

	// Extension to add in order to defer the job to the large prover
	LargeSuffix = "large"

	// Limitless prover stuffs
	BootstrapPartialSucessSuffix = "inprogress.conglomerator"
	WitnessDir                   = "/tmp/witness"
)

var (
	ALL_MODULES = [10]string{
		"ARITH-OPS", "ELLIPTIC_CURVES", "HUB-KECCAK",
		"SHA2", "ECDSA", "MODEXP_256", "G2_CHECK",
		"STATIC", "TINY-STUFFS", "MODEXP_4096",
	}
)
