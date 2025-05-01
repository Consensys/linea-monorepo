package distributed

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/cleanup"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/logdata"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mpts"
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
	fixedNbRowPlonkCircuit   = 1 << 20
	fixedNbRowExternalHasher = 1 << 16
	verifyingKeyPublicInput  = "VERIFYING_KEY"
	verifyingKey2PublicInput = "VERIFYING_KEY_2"
	lppMerkleRootPublicInput = "LPP_COLUMNS_MERKLE_ROOTS"
	// thresholdStopping self-recursion is the number of committed cells at
	// which we stop iterating the self-recursion function. It's a purely
	// empirical value and is obtained by looking at the smallest wizard
	// size that we obtain by recursing infinitely the self-recursion
	// procedure.
	thresholdStoppingSelfrecursion = 13500000

	// initialCompilerSize sets the target number of rows of the first invokation
	// of [compiler.Arcane] of the pre-recursion pass of [CompileSegment]. It is
	// also the length of the column in the [DefaultModule].
	initialCompilerSize int = 1 << 17
)

var (
	// numColumnProfileMpts tells the last invokation of Vortex prior to the self-
	// recursion to use a plonk circuit with a fixed number of rows. The values
	// are completely empirical and set to make the compilation work.
	numColumnProfileMpts            = []int{68, 1409, 147, 5, 9, 7, 0, 1}
	numColumnProfileMptsPrecomputed = 75
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
	// ModuleDefault is optional and is set if the segment is default module
	// segment.
	DefaultModule *DefaultModule
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
		modIOP            *wizard.CompiledIOP
		res               = &RecursedSegmentCompilation{}
		numActualLppRound = 0
		isLPP             bool
		subscript         string
	)

	switch m := mod.(type) {
	case *ModuleGL:
		modIOP = m.Wiop
		res.ModuleGL = m
		subscript = string(m.definitionInput.ModuleName)
	case *ModuleLPP:
		modIOP = m.Wiop
		res.ModuleLPP = m
		numActualLppRound = len(m.ModuleNames())
		isLPP = true
		subscript = fmt.Sprintf("%v", m.ModuleNames())
	case *DefaultModule:
		modIOP = m.Wiop
		res.DefaultModule = m
		subscript = "default-module"
	default:
		utils.Panic("unexpected type: %T", mod)
	}

	sisInstance := ringsis.Params{LogTwoBound: 16, LogTwoDegree: 6}

	wizard.ContinueCompilation(modIOP,
		// @alex: unsure why we need to compile with MiMC since it should be done
		// pre-bootstrapping.
		mimc.CompileMiMC,
		// The reason why 1 works is because it will work for all the GL modules
		// and because the LPP module do not have Plonk-in-wizards query.
		plonkinwizard.CompileWithMinimalRound(1),
		compiler.Arcane(
			compiler.WithTargetColSize(initialCompilerSize),
			// Some precompiles modules consists of only microscropic columns and
			// a Plonk-in-wizard query for a giant circuit. The small columns are
			// connected to the rest via a lookup but we need to ensure that the
			// columns will not be turned into "PROOF" columns and be put out of
			// the LPP commitment.
			//
			// @alex:
			// It's quite a magic number but the choice to use it nonetheless is
			// because to set the optimal value, we would need a feature where
			// the Arcane compiler detects the smallest committed column at round
			// 0 (or any round from a list) and sets its size as the StitcherMinSize
			// internally.
			//
			// For now, the current solution is fine and we can update the value from
			// time to time if not too frequent.
			compiler.WithStitcherMinSize(1<<4),
		),
	)

	if !isLPP {

		wizard.ContinueCompilation(modIOP,
			vortex.Compile(
				2,
				vortex.ForceNumOpenedColumns(256),
				vortex.WithSISParams(&sisInstance),
				vortex.AddMerkleRootToPublicInputs(lppMerkleRootPublicInput, []int{0}),
			),
		)

		for i := 1; i < lppGroupingArity; i++ {
			modIOP.InsertPublicInput(fmt.Sprintf("%v_%v", lppMerkleRootPublicInput, i), accessors.NewConstant(field.Zero()))
		}

	} else {

		wizard.ContinueCompilation(modIOP,
			vortex.Compile(
				2,
				vortex.ForceNumOpenedColumns(256),
				vortex.WithSISParams(&sisInstance),
				vortex.AddMerkleRootToPublicInputs(lppMerkleRootPublicInput, utils.RangeSlice(numActualLppRound, 0)),
			),
		)

		for i := numActualLppRound; i < lppGroupingArity; i++ {
			modIOP.InsertPublicInput(fmt.Sprintf("%v_%v", lppMerkleRootPublicInput, i), accessors.NewConstant(field.Zero()))
		}
	}

	wizard.ContinueCompilation(modIOP,
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<15),
		),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
		),
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<13),
			compiler.WithoutMpts(),
		),
	)

	// This optional step is to ensure the tightness of the final wizard by
	// adding an optional second layer of compilation when we have very
	// large inputs.
	stats := logdata.GetWizardStats(modIOP)
	if stats.NumCellsCommitted > thresholdStoppingSelfrecursion {
		wizard.ContinueCompilation(modIOP,
			mpts.Compile(),
			vortex.Compile(
				8,
				vortex.ForceNumOpenedColumns(64),
				vortex.WithSISParams(&sisInstance),
				vortex.PremarkAsSelfRecursed(),
			),
			selfrecursion.SelfRecurse,
			cleanup.CleanUp,
			mimc.CompileMiMC,
			compiler.Arcane(
				compiler.WithTargetColSize(1<<13),
				compiler.WithoutMpts(),
			),
		)
	}

	wizard.ContinueCompilation(modIOP,
		logdata.Log("just-before-recursion"),
		mpts.Compile(mpts.WithNumColumnProfileOpt(numColumnProfileMpts, numColumnProfileMptsPrecomputed)),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
			vortex.PremarkAsSelfRecursed(),
			vortex.AddPrecomputedMerkleRootToPublicInputs(verifyingKeyPublicInput),
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
				WithExternalHasherOpts: true,
				ExternalHasherNbRows:   fixedNbRowExternalHasher,
				Subscript:              subscript,
			},
		)
	}

	recursedComp := wizard.Compile(defineRecursion,
		mimc.CompileMiMC,
		plonkinwizard.Compile,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<15),
		),
		logdata.Log("just-after-recursion-expanded"),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
			vortex.AddPrecomputedMerkleRootToPublicInputs(verifyingKey2PublicInput),
		),
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<13),
		),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
		),
		selfrecursion.SelfRecurse,
		cleanup.CleanUp,
		mimc.CompileMiMC,
		compiler.Arcane(
			compiler.WithTargetColSize(1<<13),
		),
		vortex.Compile(
			8,
			vortex.ForceNumOpenedColumns(64),
			vortex.WithSISParams(&sisInstance),
			vortex.PremarkAsSelfRecursed(),
		),
	)

	res.Recursion = recCtx
	res.RecursionComp = recursedComp

	// It is necessary to add the extradata from the compiled IOP to the
	// recursed one otherwise, it will not be found.
	res.RecursionComp.ExtraData[verifyingKeyPublicInput] = modIOP.ExtraData[verifyingKeyPublicInput]

	return res
}

// ProveSegment runs the prover for a segment of the protocol
func (r *RecursedSegmentCompilation) ProveSegment(wit any) *wizard.ProverRuntime {

	var (
		comp        *wizard.CompiledIOP
		proverStep  wizard.MainProverStep
		moduleName  any
		moduleIndex int
	)

	switch m := wit.(type) {
	case *ModuleWitnessLPP:
		comp = r.ModuleLPP.Wiop
		proverStep = r.ModuleLPP.GetMainProverStep(m)
		moduleName = m.ModuleNames
		moduleIndex = m.ModuleIndex
	case *ModuleWitnessGL:
		comp = r.ModuleGL.Wiop
		proverStep = r.ModuleGL.GetMainProverStep(m)
		moduleName = m.ModuleName
		moduleIndex = m.ModuleIndex
	case nil:
		if r.DefaultModule == nil {
			utils.Panic("witness is nil but module is not default")
		}
		comp = r.DefaultModule.Wiop
		proverStep = r.DefaultModule.Assign
		moduleName = "default-module"
		moduleIndex = 0
	default:
		utils.Panic("unexpected type")
	}

	var (
		// In order to work, the recursion circuit needs to access the query params
		stoppingRound = recursion.VortexQueryRound(comp) + 1
		proverRun     *wizard.ProverRuntime
		initialTime   = profiling.TimeIt(func() {
			proverRun = wizard.RunProverUntilRound(comp, proverStep, stoppingRound)
		})
		initialProof    = proverRun.ExtractProof()
		initialProofErr = wizard.VerifyUntilRound(comp, initialProof, stoppingRound)
	)

	if initialProofErr != nil {
		panic(initialProofErr)
	}

	var (
		recStoppingRound = recursion.VortexQueryRound(r.RecursionComp) + 1
		recursionWit     = recursion.ExtractWitness(proverRun)
		run              *wizard.ProverRuntime
		recursionTime    = profiling.TimeIt(func() {
			run = wizard.RunProverUntilRound(
				r.RecursionComp,
				r.Recursion.GetMainProverStep([]recursion.Witness{recursionWit}, nil),
				recStoppingRound,
			)
		})
		finalProof    = run.ExtractProof()
		finalProofErr = wizard.VerifyUntilRound(r.RecursionComp, finalProof, recStoppingRound)
	)

	if finalProofErr != nil {
		panic(finalProofErr)
	}

	logrus.
		WithField("moduleName", moduleName).
		WithField("moduleIndex", moduleIndex).
		WithField("initial-time", initialTime).
		WithField("recursion-time", recursionTime).
		WithField("segment-type", fmt.Sprintf("%T", wit)).
		Infof("Ran prover segment")

	return run
}
