package global

import "github.com/consensys/linea-monorepo/prover-ray/wiop"

// ModuleSizeExport describes whether a compiled global verifier module has a
// statically-known domain size or expects the runtime to provide one.
type ModuleSizeExport struct {
	Dynamic    bool
	StaticSize int
}

// VerifierBucketExport is the stable, read-only view of one quotient bucket
// needed by verifier-ray code generation.
type VerifierBucketExport struct {
	Ratio          int
	Vanishings     []*wiop.Vanishing
	QuotientClaims []*wiop.Cell
}

// VerifierExport is the stable, read-only view of the compiled global verifier
// metadata needed by verifier-ray code generation.
type VerifierExport struct {
	Module        *wiop.Module
	ModuleSize    ModuleSizeExport
	MergeCoin     *wiop.CoinField
	EvalCoin      *wiop.CoinField
	WitnessViews  []*wiop.ColumnView
	WitnessClaims []*wiop.Cell
	Buckets       []VerifierBucketExport
}

// Export returns a shallow immutable snapshot of the compiled verifier metadata.
// The returned pointers identify prover-ray WIOP objects; callers must treat
// them as read-only.
func (gv *Verifier) Export() VerifierExport {
	moduleSize := ModuleSizeExport{Dynamic: gv.m.IsDynamic()}
	if !moduleSize.Dynamic {
		moduleSize.StaticSize = gv.m.Size()
	}

	buckets := make([]VerifierBucketExport, len(gv.Buckets))
	for i, bucket := range gv.Buckets {
		buckets[i] = VerifierBucketExport{
			Ratio:          bucket.Ratio,
			Vanishings:     append([]*wiop.Vanishing(nil), bucket.Vanishings...),
			QuotientClaims: append([]*wiop.Cell(nil), bucket.QuotientClaims...),
		}
	}

	return VerifierExport{
		Module:        gv.m,
		ModuleSize:    moduleSize,
		MergeCoin:     gv.MergeCoin,
		EvalCoin:      gv.EvalCoin,
		WitnessViews:  append([]*wiop.ColumnView(nil), gv.WitnessViews...),
		WitnessClaims: append([]*wiop.Cell(nil), gv.WitnessClaims...),
		Buckets:       buckets,
	}
}
