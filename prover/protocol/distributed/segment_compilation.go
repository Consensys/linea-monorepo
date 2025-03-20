package experiment

import (
	"encoding/json"

	"github.com/consensys/linea-monorepo/prover/backend/files"
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
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
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

	logCellCount(mod.Wiop, string(mod.definitionInput.ModuleName), "[ModuleLPP before compilation]")

	wizard.ContinueCompilation(mod.Wiop,
		mimc.CompileMiMC,
		plonkinwizard.Compile,
		compiler.Arcane(256, 1<<17, false),
		logdata.GenCSV(files.MustOverwrite("wizard-initial-lpp-"+string(mod.definitionInput.ModuleName)+".csv")),
		vortex.Compile(
			2,
			vortex.ForceNumOpenedColumns(256),
			vortex.WithSISParams(&sisInstance),
			vortex.AddMerkleRootToPublicInputs("LPP_COLUMNS_MERKLE_ROOT", 0),
		),
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(256, 1<<15, false),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
		),
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(256, 1<<13, false),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
			vortex.PremarkAsSelfRecursed(),
		),
	)

	logCellCount(mod.Wiop, string(mod.definitionInput.ModuleName), "[ModuleLPP before recursion]")

	var recCtx *recursion.Recursion

	defineRecursion := func(build2 *wizard.Builder) {
		recCtx = recursion.DefineRecursionOf(
			"wizard-recursion-lpp-"+string(mod.definitionInput.ModuleName),
			build2.CompiledIOP,
			mod.Wiop,
			true,
			1,
		)
	}

	recursedComp := wizard.Compile(defineRecursion,
		logdata.Log("recursion-lpp-initial-wizard-"+string(mod.definitionInput.ModuleName)),
		mimc.CompileMiMC,
		plonkinwizard.Compile,
		compiler.Arcane(256, 1<<17, false),
		logdata.GenCSV(files.MustOverwrite("wizard-recursion-lpp-"+string(mod.definitionInput.ModuleName)+".csv")),
		vortex.Compile(
			2,
			vortex.ForceNumOpenedColumns(256),
			vortex.WithSISParams(&sisInstance),
			vortex.AddMerkleRootToPublicInputs("LPP_COLUMNS_MERKLE_ROOT", 0),
		),
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(256, 1<<15, false),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
		),
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(256, 1<<13, false),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
			vortex.PremarkAsSelfRecursed(),
		),
	)

	logCellCount(recursedComp, string(mod.definitionInput.ModuleName), "[ModuleLPP after recursion]")

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

	logCellCount(mod.Wiop, string(mod.definitionInput.ModuleName), "[moduleGL before compilation]")

	wizard.ContinueCompilation(mod.Wiop,
		mimc.CompileMiMC,
		plonkinwizard.Compile,
		compiler.Arcane(256, 1<<17, false),
		logdata.GenCSV(files.MustOverwrite("wizard-initial-gl-"+string(mod.definitionInput.ModuleName)+".csv")),
		vortex.Compile(
			2,
			vortex.ForceNumOpenedColumns(256),
			vortex.WithSISParams(&sisInstance),
			vortex.AddMerkleRootToPublicInputs("LPP_COLUMNS_MERKLE_ROOT", 0),
		),
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(256, 1<<15, false),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
		),
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(256, 1<<13, false),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
			vortex.PremarkAsSelfRecursed(),
		),
	)

	logCellCount(mod.Wiop, string(mod.definitionInput.ModuleName), "[ModuleGL before recursion]")

	var recCtx *recursion.Recursion

	defineRecursion := func(build2 *wizard.Builder) {
		recCtx = recursion.DefineRecursionOf(
			"wizard-recursion-gl-"+string(mod.definitionInput.ModuleName),
			build2.CompiledIOP,
			mod.Wiop,
			true,
			1,
		)
	}

	recursedComp := wizard.Compile(defineRecursion,
		logdata.Log("recursion-gl-initial-wizard-"+string(mod.definitionInput.ModuleName)),
		mimc.CompileMiMC,
		plonkinwizard.Compile,
		compiler.Arcane(256, 1<<17, false),
		logdata.GenCSV(files.MustOverwrite("wizard-recursion-gl-"+string(mod.definitionInput.ModuleName)+".csv")),
		vortex.Compile(
			2,
			vortex.ForceNumOpenedColumns(256),
			vortex.WithSISParams(&sisInstance),
			vortex.AddMerkleRootToPublicInputs("LPP_COLUMNS_MERKLE_ROOT", 0),
		),
		selfrecursion.SelfRecurse,
		mimc.CompileMiMC,
		compiler.Arcane(256, 1<<15, false),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
			vortex.AddPrecomputedMerkleRootToPublicInputs("VERIFICATION_KEY_MERKLE_ROOT"),
		),
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(256, 1<<13, false),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
		),
	)

	logCellCount(recursedComp, string(mod.definitionInput.ModuleName), "[ModuleGL] after recursion")

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
		comp = r.ModuleGL.Wiop
		proverStep = r.ModuleGL.GetMainProverStep(wit)
	}

	var (
		// In order to work, the recursion circuit needs to access the query params
		stoppingRound = recursion.VortexQueryRound(comp) + 1
		proverRun     *wizard.ProverRuntime
		initialTime   = profiling.TimeIt(func() {
			proverRun = wizard.RunProverUntilRound(comp, proverStep, stoppingRound)
		})
		recursionWit  = recursion.ExtractWitness(proverRun)
		proof         wizard.Proof
		recursionTime = profiling.TimeIt(func() {
			proof = wizard.Prove(
				r.RecursionComp,
				r.Recursion.GetMainProverStep([]recursion.Witness{recursionWit}),
			)
		})
	)

	logrus.
		WithField("moduleName", wit.ModuleName).
		WithField("moduleIndex", wit.ModuleIndex).
		WithField("initial-time", initialTime).
		WithField("recursion-time", recursionTime).
		WithField("is-lpp", wit.IsLPP).
		Infof("Ran prover segment")

	return proof
}

func logCellCount(comp *wizard.CompiledIOP, moduleName, msg string) {
	cellCount := logdata.GetWizardStats(comp)
	cellCountJson, _ := json.Marshal(cellCount)
	logrus.Infof("[wizard.analytic] msg=%v module=%v cellCount=%v\n", msg, moduleName, string(cellCountJson))
}
