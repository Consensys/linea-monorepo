package distributed

import (
	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/cleanup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/plonkinwizard"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/recursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/selfrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/consensys/linea-monorepo/prover/utils/profiling"
	"github.com/sirupsen/logrus"
)

const (
	// fixedNbRowPlonkCircuit is the number of rows in the plonk circuit,
	// the value is empirical and corresponds to the lowest value that works.
	fixedNbRowPlonkCircuit = 1 << 24
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

// CompileSegment applies all the compilation steps required to compile an LPP
// or a GL module of the protocol. The function accepts either a *[ModuleLPP]
// or a *[ModuleGL].
func CompileSegment(mod any) *RecursedSegmentCompilation {

	var (
		modIOP *wizard.CompiledIOP
		res    = &RecursedSegmentCompilation{}
	)

	switch m := mod.(type) {
	case *ModuleGL:
		modIOP = m.Wiop
		res.ModuleGL = m
	case *ModuleLPP:
		modIOP = m.Wiop
		res.ModuleLPP = m
	default:
		utils.Panic("unexpected type: %T", mod)
	}

	sisInstance := ringsis.Params{LogTwoBound: 16, LogTwoDegree: 6}

	wizard.ContinueCompilation(modIOP,
		mimc.CompileMiMC,
		plonkinwizard.Compile,
		compiler.Arcane(256, 1<<17, false),
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

	var recCtx *recursion.Recursion

	defineRecursion := func(build2 *wizard.Builder) {
		recCtx = recursion.DefineRecursionOf(
			build2.CompiledIOP,
			modIOP,
			recursion.Parameters{
				Name:                   "wizard-recursion",
				WithoutGkr:             true,
				MaxNumProof:            1,
				FixedNbRowPlonkCircuit: fixedNbRowPlonkCircuit,
			},
		)
	}

	recursedComp := wizard.Compile(defineRecursion,
		mimc.CompileMiMC,
		plonkinwizard.Compile,
		compiler.Arcane(256, 1<<17, false),
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

	res.Recursion = recCtx
	res.RecursionComp = recursedComp
	return res
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
