package compiler

import (
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/cleanup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/globalcs"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/horner"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/innerproduct"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/localcs"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logdata"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logderivativesum"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mpts"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/permutation"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/specialqueries"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/stitchsplit"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/univariates"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// ArcaneParams is an option for the Arcane compiler
type ArcaneParams func(*arcaneParamSet)

// WithStitcherMinSize sets the minimum size for the stitcher. All columns
// under this size are moved to public columns.
func WithStitcherMinSize(minStickSize int) ArcaneParams {
	return func(set *arcaneParamSet) {
		set.minStickSize = minStickSize
	}
}

// WithTargetColSize sets the target size for the columns.
func WithTargetColSize(targetColSize int) ArcaneParams {
	return func(set *arcaneParamSet) {
		set.targetColSize = targetColSize
	}
}

// WithLogs tells the compiler to logs compilation stats.
func WithLogs() ArcaneParams {
	return func(set *arcaneParamSet) {
		set.withLogs = true
	}
}

// WithoutMpts tells the compiler to skip the Mpts compiler.
func WithoutMpts() ArcaneParams {
	return func(set *arcaneParamSet) {
		set.WithoutMpts = true
	}
}

// arcaneParamSet collects optional parameters for the Arcane compiler.
type arcaneParamSet struct {
	minStickSize  int
	targetColSize int
	withLogs      bool
	WithoutMpts   bool
}

// Arcane is a grouping of all compilers. It compiles
// any wizard into a single-point polynomial-IOP
func Arcane(options ...ArcaneParams) func(comp *wizard.CompiledIOP) {

	params := &arcaneParamSet{}
	for _, op := range options {
		op(params)
	}

	if params.minStickSize == 0 {
		params.minStickSize = 256
	}

	return func(comp *wizard.CompiledIOP) {
		if params.withLogs {
			logdata.Log("initially")(comp)
		}

		specialqueries.RangeProof(comp)
		specialqueries.CompileFixedPermutations(comp)
		permutation.CompileViaGrandProduct(comp)
		logderivativesum.CompileLookups(comp)
		horner.CompileProjection(comp)
		innerproduct.Compile(comp)

		if params.withLogs {
			logdata.Log("after-expansion")(comp)
		}

		stitchsplit.Stitcher(params.minStickSize, params.targetColSize)(comp)
		stitchsplit.Splitter(params.targetColSize)(comp)

		if params.withLogs {
			logdata.Log("post-rectangularization")(comp)
		}

		cleanup.CleanUp(comp)
		localcs.Compile(comp)
		globalcs.Compile(comp)
		univariates.CompileLocalOpening(comp)
		univariates.Naturalize(comp)

		if !params.WithoutMpts {
			mpts.Compile()(comp)
		}

		if params.withLogs {
			logdata.Log("end-of-arcane")(comp)
		}
	}
}
