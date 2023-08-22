package plonk

// PlonkOption allows modifying Plonk circuit arithmetization.
type Option func(*Ctx)

// WithRangecheck allows bridging range checking from gnark into Wizard. The
// total of bits being range-checked are nbBits*nbLimbs.
func WithRangecheck(nbBits, nbLimbs int) Option {
	return func(c *Ctx) {
		c.RangeCheck.Enabled = true
		c.RangeCheck.NbBits = nbBits
		c.RangeCheck.NbLimbs = nbLimbs
	}
}
