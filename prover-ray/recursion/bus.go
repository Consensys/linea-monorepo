package recursion

import (
	"github.com/LFDT-Lineth/lineth-monorepo/prover-ray/wiop"
	"github.com/LFDT-Lineth/lineth-monorepo/prover-ray/wiop/compilers/lookuptologderivsum"
)

// access is one memory access of a chip row: the tuple (addr, v0..v7) with
// multiplicity mult — the number of reads of the cell for a write, −1 times a
// selector for a read, 0 on rows that do not access.
type access struct {
	mod  *module
	addr wiop.Expression
	vals [CellWidth]wiop.Expression
	mult wiop.Expression
}

// Bus is the write-once memory: every address is written by exactly one chip
// row, with its read count as multiplicity, and each read is a −1 term. The
// accesses of all chips balance in one log-derivative sum:
//
//	Σ_writes mult/(γ + fold_α(addr, v)) − Σ_reads 1/(γ + fold_α(addr, v)) = 0,
//
// fold_α(addr, v) = addr + α·v0 + … + α^8·v7, α and γ drawn after the
// chips' columns are committed. The program assigns each address to a single
// write and counts its reads, and both are setup columns, so a read can only
// be balanced by the write of its address, with the same value.
type Bus struct {
	accesses []access
}

// Write adds a write of vals at addr with multiplicity mult.
func (b *Bus) Write(m *module, addr wiop.Expression, vals [CellWidth]wiop.Expression, mult wiop.Expression) {
	b.accesses = append(b.accesses, access{m, addr, vals, mult})
}

// Read adds a read of vals at addr, active when sel is 1.
func (b *Bus) Read(m *module, addr wiop.Expression, vals [CellWidth]wiop.Expression, sel wiop.Expression) {
	b.accesses = append(b.accesses, access{m, addr, vals, wiop.Negate(sel)})
}

// build registers the bus on sys: the coins α and γ in coinRound (after the
// round of the chips' columns), one fraction mult/(γ + fold_α(addr, v)) per
// access, their log-derivative sum, and a verifier action asserting that it
// is zero (lookuptologderivsum's, which the verifier-ray codegen recognizes).
// The logderivativesum pass then packs the fractions of each module into its
// running-sum columns.
func (b *Bus) build(sys *wiop.System, coinRound *wiop.Round) *wiop.LogDerivativeSum {
	if len(b.accesses) == 0 {
		return nil
	}
	ctx := sys.Context.Childf("bus")
	alpha := coinRound.NewCoinField(ctx.Childf("alpha"))
	gamma := coinRound.NewCoinField(ctx.Childf("gamma"))
	fractions := make([]wiop.Fraction, len(b.accesses))
	for i, a := range b.accesses {
		tuple := append([]wiop.Expression{a.addr}, a.vals[:]...)
		fractions[i] = wiop.Fraction{
			Numerator:   a.mult,
			Denominator: wiop.Add(gamma, wiop.RLCExpression(alpha, tuple)),
		}
	}
	ld := sys.NewLogDerivativeSum(ctx.Childf("sum"), fractions)
	ld.Result.Round().RegisterVerifierAction(&lookuptologderivsum.ResultIsZeroVerifierAction{LogDerivativeSum: ld})
	return ld
}
