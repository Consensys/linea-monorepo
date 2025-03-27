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

// WithFixedNbRows fixes the number of rows to allocate in the Plonk columns.
// Without the option, the number of rows is the next power of two of the
// number of constraints. The option overrides it. However, the provided
// number of rows must be higher than the number of constraints of the
// circuit otherwise, the compilation will fail with panic.
func WithFixedNbRows(nbRow int) Option {
	return func(c *CompilationCtx) {
		c.FixedNbRowsOption.Enabled = true
		c.FixedNbRowsOption.NbRow = nbRow
	}
}
