package distributed

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/crypto/mimc"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/query"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

var (
	logDerivativeSumPublicInput = "LOG_DERIVATE_SUM_PUBLIC_INPUT"
	grandProductPublicInput     = "GRAND_PRODUCT_PUBLIC_INPUT"
	hornerPublicInput           = "HORNER_FINAL_RES_PUBLIC_INPUT"
	hornerN0HashPublicInput     = "HORNER_N0_HASH_PUBLIC_INPUT"
	hornerN1HashPublicInput     = "HORNER_N1_HASH_PUBLIC_INPUT"
	isLppPublicInput            = "IS_LPP"
	isGlPublicInput             = "IS_GL"
	nbActualLppPublicInput      = "NB_ACTUAL_LPP"
)

// ModuleLPP is a compilation structure holding the central informations
// of the LPP part of a module.
type ModuleLPP struct {

	// moduleTranslator is the translator for the GL part of the module
	// it also has the ownership of the [wizard.Compiled] IOP built for
	// this module.
	moduleTranslator

	// definitionInputs stores the [FilteredModuleInputs] that was used
	// to generate the module.
	definitionInputs []FilteredModuleInputs

	// InitialFiatShamirState is the state at which to start the FiatShamir
	// computation
	InitialFiatShamirState ifaces.Column

	// N0Hash is the hash of the N0 positions for the Horner queries
	N0Hash ifaces.Column

	// N1Hash is the hash of the N1 positions for the Horner queries
	N1Hash ifaces.Column

	// LogDerivativeSum is the translated log-derivative query in the module
	LogDerivativeSum query.LogDerivativeSum

	// GrandProduct is the translated grand-product query in the module
	GrandProduct query.GrandProduct

	// Horner is the translated horner query in the module
	Horner query.Horner
}

// SetInitialFSHash sets the initial FiatShamir state
type SetInitialFSHash struct {
	ModuleLPP
	skipped bool
}

// AssignLPPQueries is a [wizard.ProverAction] responsible for assigning the LPP
// query results as well as the N0Hash and N1Hash values.
type AssignLPPQueries struct {
	ModuleLPP
}

// CheckNxHash is a [wizard.VerifierAction] responsible for checking the N0Hash
// and N1Hash values.
type CheckNxHash struct {
	ModuleLPP
	skipped bool
}

// LppWitnessAssignment is a [wizard.ProverAction] responsible for assigning the
// LPP witness values at round "round".
type LppWitnessAssignment struct {
	ModuleLPP
	round int
}

// BuildModuleLPP builds a [ModuleLPP] from scratch from a [FilteredModuleInputs].
// The function works by creating a define function that will call [NewModuleLPP]
// / and then it calls [wizard.Compile] without passing compilers.
func BuildModuleLPP(moduleInput []FilteredModuleInputs) *ModuleLPP {

	var (
		moduleLPP  *ModuleLPP
		defineFunc = func(b *wizard.Builder) {
			moduleLPP = NewModuleLPP(b, moduleInput)
		}
		// Since the NewModuleLPP contains a pointer to the compiled IOP already
		// there is no need to use the one returned by [wizard.Compile].
		_ = wizard.Compile(defineFunc)
	)

	return moduleLPP
}

// NewModuleLPP declares and constructs a new [ModuleLPP] from a [wizard.Builder]
// and a [FilteredModuleInput]. The function performs all the necessary
// declarations to build the LPP part of the module and returns the constructed
// module.
func NewModuleLPP(builder *wizard.Builder, moduleInputs []FilteredModuleInputs) *ModuleLPP {

	moduleLPP := &ModuleLPP{
		moduleTranslator: moduleTranslator{
			Wiop: builder.CompiledIOP,
			Disc: moduleInputs[0].Disc,
		},
		definitionInputs:       moduleInputs,
		InitialFiatShamirState: builder.InsertProof(0, "INITIAL_FIATSHAMIR_STATE", 1),
	}

	// The starting round is the round where we can add data other than the LPP
	// columns.
	var startingRound = len(moduleInputs)

	// These are the "dummy" public inputs that are only here so that the
	// moduleGL and moduleLPP have identical set of public inputs. The order
	// of declaration is also important. Namely, these needs to be declared before
	// the non-dummy ones.
	moduleLPP.Wiop.InsertPublicInput(isFirstPublicInput, accessors.NewConstant(field.Zero()))
	moduleLPP.Wiop.InsertPublicInput(isLastPublicInput, accessors.NewConstant(field.Zero()))
	moduleLPP.Wiop.InsertPublicInput(globalReceiverPublicInput, accessors.NewConstant(field.Zero()))
	moduleLPP.Wiop.InsertPublicInput(globalSenderPublicInput, accessors.NewConstant(field.Zero()))

	for round, moduleInput := range moduleInputs {
		for _, col := range moduleInput.Columns {

			if col.Round() != 0 {
				utils.Panic("cannot translate a column with non-zero round %v", col.Round())
			}

			_, isLPP := moduleInput.ColumnsLPPSet[col.GetColID()]

			if !isLPP {
				continue
			}

			if data, isPrecomp := moduleInput.ColumnsPrecomputed[col.GetColID()]; isPrecomp {
				moduleLPP.InsertPrecomputed(*col, data)
				continue
			}

			moduleLPP.InsertColumn(*col, round)
		}
	}

	logDerivativeArgs, grandProductArgs, hornerArgs := getQueryArgs(moduleInputs)

	if len(logDerivativeArgs) > 0 {

		moduleLPP.LogDerivativeSum = moduleLPP.InsertLogDerivative(
			startingRound,
			ifaces.QueryID("MAIN_LOGDERIVATIVE"),
			logDerivativeArgs,
		)

		moduleLPP.Wiop.InsertPublicInput(
			logDerivativeSumPublicInput,
			accessors.NewLogDerivSumAccessor(moduleLPP.LogDerivativeSum),
		)

	} else {

		moduleLPP.Wiop.InsertPublicInput(
			logDerivativeSumPublicInput,
			accessors.NewConstant(field.Zero()),
		)
	}

	if len(grandProductArgs) > 0 {

		moduleLPP.GrandProduct = moduleLPP.InsertGrandProduct(
			startingRound,
			ifaces.QueryID("MAIN_GRANDPRODUCT"),
			grandProductArgs,
		)

		moduleLPP.Wiop.InsertPublicInput(
			grandProductPublicInput,
			accessors.NewGrandProductAccessor(moduleLPP.GrandProduct),
		)

	} else {

		moduleLPP.Wiop.InsertPublicInput(
			grandProductPublicInput,
			accessors.NewConstant(field.One()),
		)
	}

	if len(hornerArgs) > 0 {

		moduleLPP.Horner = moduleLPP.InsertHorner(
			startingRound,
			ifaces.QueryID("MAIN_HORNER"),
			hornerArgs,
		)

		moduleLPP.N0Hash = builder.InsertProof(startingRound, "LPP_N0_HASH", 1)
		moduleLPP.N1Hash = builder.InsertProof(startingRound, "LPP_N1_HASH", 1)

		moduleLPP.Wiop.InsertPublicInput(
			hornerPublicInput,
			accessors.NewFromHornerAccessorFinalValue(&moduleLPP.Horner),
		)

		moduleLPP.Wiop.InsertPublicInput(
			hornerN0HashPublicInput,
			accessors.NewFromPublicColumn(moduleLPP.N0Hash, 0),
		)

		moduleLPP.Wiop.InsertPublicInput(
			hornerN1HashPublicInput,
			accessors.NewFromPublicColumn(moduleLPP.N1Hash, 0),
		)

		moduleLPP.Wiop.RegisterVerifierAction(startingRound, &CheckNxHash{ModuleLPP: *moduleLPP})

	} else {

		moduleLPP.Wiop.InsertPublicInput(
			hornerPublicInput,
			accessors.NewConstant(field.Zero()),
		)

		moduleLPP.Wiop.InsertPublicInput(
			hornerN0HashPublicInput,
			accessors.NewConstant(field.Zero()),
		)

		moduleLPP.Wiop.InsertPublicInput(
			hornerN1HashPublicInput,
			accessors.NewConstant(field.Zero()),
		)
	}

	for _, pi := range moduleInputs[0].PublicInputs {
		moduleLPP.Wiop.InsertPublicInput(pi.Name, accessors.NewConstant(field.Zero()))
	}

	moduleLPP.Wiop.InsertPublicInput(isGlPublicInput, accessors.NewConstant(field.Zero()))
	moduleLPP.Wiop.InsertPublicInput(isLppPublicInput, accessors.NewConstant(field.One()))
	moduleLPP.Wiop.InsertPublicInput(nbActualLppPublicInput, accessors.NewConstant(field.NewElement(uint64(len(moduleInputs)))))

	// In case the LPP part is empty, we have a scenario where the sub-proof to
	// build has no registered coin. This creates errors in the compilation
	// due to sanity-check firing up. We add a coin to remediate.
	for i := 0; i < len(moduleInputs); i++ {
		moduleLPP.InsertCoin(coin.Namef("LPP_DUMMY_COIN_%v", i+1), i+1)
	}

	for round := 1; round < len(moduleInputs); round++ {
		moduleLPP.Wiop.RegisterProverAction(round, LppWitnessAssignment{ModuleLPP: *moduleLPP, round: round})
	}

	moduleLPP.Wiop.RegisterProverAction(startingRound, &AssignLPPQueries{*moduleLPP})
	moduleLPP.Wiop.FiatShamirHooksPreSampling.AppendToInner(1, &SetInitialFSHash{ModuleLPP: *moduleLPP})

	return moduleLPP
}

// ModuleNames returns the list of the module names of the [ModuleLPP].
func (m *ModuleLPP) ModuleNames() []ModuleName {
	res := make([]ModuleName, 0)
	for _, definitionInput := range m.definitionInputs {
		res = append(res, definitionInput.ModuleName)
	}
	return res
}

// GetMainProverStep returns a [wizard.ProverStep] running [Assign] passing
// the provided [ModuleWitness] argument.
func (m *ModuleLPP) GetMainProverStep(witness *ModuleWitnessLPP) wizard.MainProverStep {
	return func(run *wizard.ProverRuntime) {
		m.Assign(run, witness)
	}
}

// Assign is the entry-point for the assignment of the [ModuleLPP]. It
// is responsible for setting up the [ProverRuntime.State] with the witness
// value and assigning the columns.
//
// The function will only populates the [ModuleWitness] from its columns
// assignment as the columns are assigned in the runtime. The function will
// only assign the columns corresponding to the first submodule The
// others are done via [LppWitnessAssignment] with round > 0.
func (m *ModuleLPP) Assign(run *wizard.ProverRuntime, witness *ModuleWitnessLPP) {
	run.State.InsertNew(moduleWitnessKey, witness)

	run.AssignColumn(
		m.InitialFiatShamirState.GetColID(),
		smartvectors.NewConstant(witness.InitialFiatShamirState, 1),
	)

	a := LppWitnessAssignment{ModuleLPP: *m, round: 0}
	a.Run(run)
}

func (a LppWitnessAssignment) Run(run *wizard.ProverRuntime) {

	var (
		witness = run.State.MustGet(moduleWitnessKey).(*ModuleWitnessLPP)
		m       = a.ModuleLPP
		round   = a.round
	)

	// Note @alex: It should be fine to look only at m.definitionInputs[round]
	// instead of scanning through all the definitionInputs.
	for _, definitionInput := range m.definitionInputs {

		// [definitionInput.Columns] stores the list of columns to assign.
		// Though, it stores the columns as in the origin CompiledIOP so we
		// cannot directly use them to refer to columns of the current IOP.
		// Yet, the column share the same names.
		for _, col := range definitionInput.Columns {

			colName := col.GetColID()

			// Skips the non-LPP columns
			if _, ok := definitionInput.ColumnsLPPSet[colName]; !ok {
				continue
			}

			newCol := m.Wiop.Columns.GetHandle(colName)

			if newCol.Round() != round {
				continue
			}

			if m.Wiop.Precomputed.Exists(colName) {
				continue
			}

			colWitness, ok := witness.Columns[colName]
			if !ok {
				utils.Panic("witness of column %v was not found", colName)
			}

			run.AssignColumn(colName, colWitness)
			delete(witness.Columns, colName)
		}
	}
}

// addCoinFromExpression scans the metadata of the expression looking
// for coins and adds them to the [ModuleLPP] as [coin.FieldFromSeed].
func (m *ModuleLPP) addCoinFromExpression(expr *symbolic.Expression) {

	var (
		board    = expr.Board()
		metadata = board.ListVariableMetadata()
	)

	for i := range metadata {

		switch meta := metadata[i].(type) {

		case coin.Info:

			m.InsertCoin(meta.Name, meta.Round)
			return

		case ifaces.Accessor:

			m.addCoinFromAccessor(meta)
			return
		}
	}
}

func (m *ModuleLPP) addCoinFromAccessor(acce ifaces.Accessor) {

	switch a := acce.(type) {

	case *accessors.FromExprAccessor:

		m.addCoinFromExpression(a.Expr)
		return

	case *accessors.FromCoinAccessor:

		m.InsertCoin(a.Info.Name, a.Info.Round)
		return
	}
}

func (a AssignLPPQueries) Run(run *wizard.ProverRuntime) {

	moduleWitness := run.State.MustGet(moduleWitnessKey).(*ModuleWitnessLPP)
	run.State.Del(moduleWitnessKey)

	logDerivativeArgs, grandProductArgs, hornerArgs := getQueryArgs(a.definitionInputs)

	if len(hornerArgs) > 0 {
		hornerParams := a.getHornerParams(run, moduleWitness.N0Values)
		run.AssignHornerParams(a.Horner.ID, hornerParams)
		n0Hash, n1Hash := hashNxs(hornerParams, 0), hashNxs(hornerParams, 1)

		run.AssignColumn(a.N0Hash.GetColID(), smartvectors.NewRegular([]field.Element{n0Hash}))
		run.AssignColumn(a.N1Hash.GetColID(), smartvectors.NewRegular([]field.Element{n1Hash}))
	}

	if len(grandProductArgs) > 0 {
		run.AssignGrandProduct(a.GrandProduct.ID, a.GrandProduct.Compute(run))
	}

	if len(logDerivativeArgs) > 0 {

		y, err := a.LogDerivativeSum.Compute(run)
		if err != nil {
			utils.Panic("LogDerivative has a zero term in the denominator: %v", err)
		}

		run.AssignLogDerivSum(a.LogDerivativeSum.ID, y)
	}
}

func (m ModuleLPP) getHornerParams(run *wizard.ProverRuntime, n0Values []int) query.HornerParams {

	hornerParams := query.HornerParams{}
	for i := range n0Values {
		hornerParams.Parts = append(hornerParams.Parts, query.HornerParamsPart{
			N0: n0Values[i],
		})
	}

	hornerParams.SetResult(run, m.Horner)
	return hornerParams
}

func (a *CheckNxHash) Run(run wizard.Runtime) error {

	var (
		hornerParams  = run.GetHornerParams(a.Horner.ID)
		n0HashAlleged = a.N0Hash.GetColAssignmentAt(run, 0)
		n1HashAlleged = a.N1Hash.GetColAssignmentAt(run, 0)
		n0Hash        = hashNxs(hornerParams, 0)
		n1Hash        = hashNxs(hornerParams, 1)
	)

	if n1HashAlleged != n1Hash {
		return fmt.Errorf("n0Hash %v != n1HashAlleged %v", n1Hash, n1HashAlleged)
	}

	if n0HashAlleged != n0Hash {
		return fmt.Errorf("n0Hash %v != n0HashAlleged %v", n0Hash, n0HashAlleged)
	}

	return nil
}

func (a *CheckNxHash) RunGnark(api frontend.API, run wizard.GnarkRuntime) {

	var (
		hornerParams  = run.GetHornerParams(a.Horner.ID)
		n0HashAlleged = a.N0Hash.GetColAssignmentGnarkAt(run, 0)
		n1HashAlleged = a.N1Hash.GetColAssignmentGnarkAt(run, 0)
		n0Hash        = hashNxsGnark(run.GetHasherFactory(), hornerParams, 0)
		n1Hash        = hashNxsGnark(run.GetHasherFactory(), hornerParams, 1)
	)

	api.AssertIsEqual(n0Hash, n0HashAlleged)
	api.AssertIsEqual(n1Hash, n1HashAlleged)
}

func (a *CheckNxHash) Skip() {
	a.skipped = true
}

func (a *CheckNxHash) IsSkipped() bool {
	return a.skipped
}

func (a *SetInitialFSHash) Run(run wizard.Runtime) error {
	state := a.InitialFiatShamirState.GetColAssignment(run).Get(0)
	run.Fs().SetState([]field.Element{state})
	return nil
}

func (a *SetInitialFSHash) RunGnark(api frontend.API, run wizard.GnarkRuntime) {
	state := a.InitialFiatShamirState.GetColAssignmentGnark(run)[0]
	run.Fs().SetState([]frontend.Variable{state})
}

func (a *SetInitialFSHash) Skip() {
	a.skipped = true
}

func (a *SetInitialFSHash) IsSkipped() bool {
	return a.skipped
}

// hashNxs scans params and hash either the N0s or the N1s value all together
// (pass x=0, to compute the hash of the N0s and x=1 for the N1s).
func hashNxs(params query.HornerParams, x int) field.Element {

	hsh := mimc.NewMiMC()

	for _, part := range params.Parts {

		nx := 0

		if x == 0 {
			nx = part.N0
		} else {
			nx = part.N1
		}

		nxField := field.NewElement(uint64(nx))
		nxBytes := nxField.Bytes()
		hsh.Write(nxBytes[:])
	}

	resBytes := hsh.Sum(nil)
	var res field.Element
	res.SetBytes(resBytes)

	return res
}

// hashNxsGnark is as [hashNxs] but in a gnark circuit
func hashNxsGnark(factory mimc.HasherFactory, params query.GnarkHornerParams, x int) frontend.Variable {

	hsh := factory.NewHasher()

	for _, part := range params.Parts {

		var nx frontend.Variable

		if x == 0 {
			nx = part.N0
		} else {
			nx = part.N1
		}

		hsh.Write(nx)
	}

	return hsh.Sum()
}

// getQueryArgs groups the args of the [FilteredModuleInputs] provided
// by the caller.
func getQueryArgs(moduleInputs []FilteredModuleInputs) (logDerivativeArgs, grandProductArgs [][2]*symbolic.Expression, hornerArgs []query.HornerPart) {
	for _, moduleInput := range moduleInputs {
		logDerivativeArgs = append(logDerivativeArgs, moduleInput.LogDerivativeArgs...)
		grandProductArgs = append(grandProductArgs, moduleInput.GrandProductArgs...)
		hornerArgs = append(hornerArgs, moduleInput.HornerArgs...)
	}
	return logDerivativeArgs, grandProductArgs, hornerArgs
}
