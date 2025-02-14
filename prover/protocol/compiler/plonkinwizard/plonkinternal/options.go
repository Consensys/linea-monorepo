package plonkinternal

// PlonkOption allows modifying Plonk circuit arithmetization.
type Option func(*CompilationCtx)

// WithRangecheck allows bridging range checking from gnark into Wizard. The
// total of bits being range-checked are nbBits*nbLimbs. If addGateForRangeCheck
// is true, then new gates are added for wires not present in existing gates.
func WithRangecheck(nbBits, nbLimbs int, addGateForRangeCheck bool) Option {
	return func(c *CompilationCtx) {
		c.RangeCheck.Enabled = true
		c.RangeCheck.NbBits = nbBits
		c.RangeCheck.NbLimbs = nbLimbs
		c.RangeCheck.AddGateForRangeCheck = addGateForRangeCheck
	}
}
