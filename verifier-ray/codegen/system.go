package codegen

// CompiledSystem bundles all sub-verifier metadata derived from a single
// wiop.System after the full compiler pipeline has run. Each field covers one
// sub-verifier; fields are zero-valued (empty) when the system has no queries
// of that kind.
type CompiledSystem struct {
	Name      string
	Routing   CoinRouting
	Vanishing VanishingSystem
	LogDeriv  LogDerivSystem
}
