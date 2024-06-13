package config

// TODO @gbotrel don't think these should be defined here, to refactor.

// logLevel defines is a type representing different logging levels.
type logLevel uint8

const (
	// Possible levels of logging.
	_logTrace logLevel = iota + 1
	_logDebug
	_logInfo
	_logWarn
	_logError
	_logFatal
)

type ProverMode string

const (
	ProverModeDev       ProverMode = "dev"
	ProverModePartial   ProverMode = "partial"
	ProverModeFull      ProverMode = "full"
	ProverModeProofless ProverMode = "proofless"
)
