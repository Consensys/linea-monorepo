package experiment

import (
	"encoding/json"

	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/cleanup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logdata"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/plonkinwizard"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/sirupsen/logrus"
)

// RecursedSegmentCompilation collects all the wizard compilation artefacts
// to compile a segment of the protocol into a standardized recursed proof.
// A standardized proof has a verifier with a standard structure which has
// the same verifier circuit as other segments of the protocol.
type RecursedSegmentCompilation struct {
	// ModuleGL is optional and is set if the segment is a GL segment.
	ModuleGL *ModuleGL
	// ModuleLPP is optional and is set if the segment is a LPP segment.
	ModuleLPP *ModuleLPP
	// Recursion is the wizard construction context of the recursed wizard.
	Recursion *recursion.Recursion
	// RecursionComp is the compiled IOP of the recursed wizard.
	RecursionComp *wizard.CompiledIOP
}

// CompileSegmentLPP applies all the compilation steps required to compile
// an LPP module of the protocol.
func CompileSegmentLPP(mod *ModuleLPP) RecursedSegmentCompilation {

	sisInstance := ringsis.Params{LogTwoBound: 16, LogTwoDegree: 6}

	wizard.ContinueCompilation(mod.Wiop,
		mimc.CompileMiMC,
		plonkinwizard.Compile,
		compiler.Arcane(1, 1<<17, true),
		vortex.Compile(
			2,
			vortex.ForceNumOpenedColumns(256),
			vortex.WithSISParams(&sisInstance),
			vortex.AddMerkleRootToPublicInputs("LPP_COLUMNS_MERKLE_ROOT", 0),
		),
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(1, 1<<15, true),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
		),
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(1, 1<<13, true),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
			vortex.PremarkAsSelfRecursed(),
		),
	)

	var recCtx *recursion.Recursion

	defineRecursion := func(build2 *wizard.Builder) {
		recCtx = recursion.DefineRecursionOf("test", build2.CompiledIOP, mod.Wiop, true, 1)
	}

	recursedComp := wizard.Compile(
		defineRecursion,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(1, 1<<15, true),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
		),
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(1, 1<<13, true),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
			vortex.PremarkAsSelfRecursed(),
		),
	)

	cellCount := logdata.CountCells(recursedComp)
	cellCountJson, _ := json.Marshal(cellCount)
	logrus.Infof("[ModuleLPP] module=%v finalCellCount=%v", mod.definitionInput.ModuleName, string(cellCountJson))

	return RecursedSegmentCompilation{
		ModuleLPP:     mod,
		Recursion:     recCtx,
		RecursionComp: recursedComp,
	}
}

// CompileSegmentGL applies all the compilation steps required to compile
// a GL module of the protocol.
func CompileSegmentGL(mod *ModuleGL) RecursedSegmentCompilation {

	sisInstance := ringsis.Params{LogTwoBound: 16, LogTwoDegree: 6}

	wizard.ContinueCompilation(mod.Wiop,
		mimc.CompileMiMC,
		plonkinwizard.Compile,
		compiler.Arcane(2, 1<<17, true),
		vortex.Compile(
			2,
			vortex.ForceNumOpenedColumns(256),
			vortex.WithSISParams(&sisInstance),
			vortex.AddMerkleRootToPublicInputs("LPP_COLUMNS_MERKLE_ROOT", 0),
		),
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(2, 1<<15, true),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
		),
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(2, 1<<13, true),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
			vortex.PremarkAsSelfRecursed(),
		),
	)

	var recCtx *recursion.Recursion

	defineRecursion := func(build2 *wizard.Builder) {
		recCtx = recursion.DefineRecursionOf("test", build2.CompiledIOP, mod.Wiop, true, 1)
	}

	recursedComp := wizard.Compile(defineRecursion,
		mimc.CompileMiMC,
		compiler.Arcane(2, 1<<15, true),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
			vortex.AddPrecomputedMerkleRootToPublicInputs("VERIFICATION_KEY_MERKLE_ROOT"),
		),
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(2, 1<<13, true),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
		),
	)

	cellCount := logdata.CountCells(recursedComp)
	cellCountJson, _ := json.Marshal(cellCount)
	logrus.Infof("[ModuleGL] module=%v cellCount=%v", mod.definitionInput.ModuleName, string(cellCountJson))

	return RecursedSegmentCompilation{
		ModuleGL:      mod,
		Recursion:     recCtx,
		RecursionComp: recursedComp,
	}
}

// ProveSegment runs the prover for a segment of the protocol
func (r *RecursedSegmentCompilation) ProveSegment(wit *ModuleWitness) wizard.Proof {

	var (
		comp       *wizard.CompiledIOP
		proverStep wizard.ProverStep
	)

	if wit.IsLPP {
		comp = r.ModuleLPP.Wiop
		proverStep = r.ModuleLPP.GetMainProverStep(wit)
	} else {
		comp = r.ModuleLPP.Wiop
		proverStep = r.ModuleGL.GetMainProverStep(wit)
	}

	var (
		stoppingRound = recursion.LastVortexCommitRound(comp)
		proverRun     = wizard.RunProverUntilRound(comp, proverStep, stoppingRound)
		recursionWit  = recursion.ExtractWitness(proverRun)
	)

	return wizard.Prove(
		r.RecursionComp,
		r.Recursion.GetMainProverStep([]recursion.Witness{recursionWit}),
	)
}
