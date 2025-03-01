package experiment

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
	hornerPublicInput           = "HORNER_PUBLIC_INPUT"
)

// ModuleLPP is a compilation structure holding the central informations
// of the LPP part of a module.
type ModuleLPP struct {

	// moduleTranslator is the translator for the GL part of the module
	// it also has the ownership of the [wizard.Compiled] IOP built for
	// this module.
	moduleTranslator

	// definitionInput stores the [FilteredModuleInputs] that was used
	// to generate the module.
	definitionInput *FilteredModuleInputs

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

// BuildModuleLPP builds a [ModuleLPP] from scratch from a [FilteredModuleInputs].
// The function works by creating a define function that will call [NewModuleLPP]
// / and then it calls [wizard.Compile] without passing compilers.
func BuildModuleLPP(moduleInput *FilteredModuleInputs) *ModuleLPP {

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
func NewModuleLPP(builder *wizard.Builder, moduleInput *FilteredModuleInputs) *ModuleLPP {

	moduleLPP := &ModuleLPP{
		moduleTranslator: moduleTranslator{
			Wiop: builder.CompiledIOP,
			Disc: moduleInput.Disc,
		},
		definitionInput:        moduleInput,
		N0Hash:                 builder.InsertProof(1, "LPP_N0_HASH", 1),
		N1Hash:                 builder.InsertProof(1, "LPP_N1_HASH", 1),
		InitialFiatShamirState: builder.InsertProof(0, "INITIAL_FIATSHAMIR_STATE", 1),
	}

	for _, col := range moduleInput.Columns {

		if col.Round() != 0 {
			utils.Panic("cannot translate a column with non-zero round %v", col.Round())
		}

		_, isLPP := moduleInput.ColumnsLPPSet[col.GetColID()]

		if !isLPP {
			continue
		}

		moduleLPP.InsertColumn(*col, 0)

		if data, isPrecomp := moduleInput.ColumnsPrecomputed[col.GetColID()]; isPrecomp {
			moduleLPP.Wiop.Precomputed.InsertNew(col.ID, data)
		}
	}

	moduleLPP.LogDerivativeSum = moduleLPP.InsertLogDerivative(1, ifaces.QueryID("MAIN_LOGDERIVATIVE"), moduleInput.LogDerivativeArgs)
	moduleLPP.GrandProduct = moduleLPP.InsertGrandProduct(1, ifaces.QueryID("MAIN_GRANDPRODUCT"), moduleInput.GrandProductArgs)
	moduleLPP.Horner = moduleLPP.InsertHorner(1, ifaces.QueryID("MAIN_HORNER"), moduleInput.HornerArgs)

	moduleLPP.Wiop.InsertPublicInput(logDerivativeSumPublicInput, accessors.NewLogDerivSumAccessor(moduleLPP.LogDerivativeSum))
	moduleLPP.Wiop.InsertPublicInput(grandProductPublicInput, accessors.NewGrandProductAccessor(moduleLPP.GrandProduct))
	moduleLPP.Wiop.InsertPublicInput(hornerPublicInput, accessors.NewFromHornerAccessorFinalValue(&moduleLPP.Horner))

	moduleLPP.Wiop.RegisterProverAction(1, &AssignLPPQueries{*moduleLPP})
	moduleLPP.Wiop.RegisterVerifierAction(1, &CheckNxHash{ModuleLPP: *moduleLPP})
	moduleLPP.Wiop.FiatShamirHooksPreSampling.AppendToInner(1, &SetInitialFSHash{ModuleLPP: *moduleLPP})

	return moduleLPP
}

// Assign is the entry-point for the assignment of the [ModuleLPP]. It
// is responsible for setting up the [ProverRuntime.State] with the witness
// value and assigning the columns.
//
// The function depopulates the [ModuleWitness] from its columns assignment
// as the columns are assigned in the runtime.
func (m *ModuleLPP) Assign(run *wizard.ProverRuntime, witness *ModuleWitness) {

	var (
		// columns stores the list of columns to assign. Though, it
		// stores the columns as in the origin CompiledIOP so we cannot
		// directly use them to refer to columns of the current IOP.
		// Yet, the column share the same names.
		columns = m.definitionInput.Columns
	)

	run.State.InsertNew(moduleWitnessKey, witness)

	for _, col := range columns {

		colName := col.GetColID()

		// Skips the non-LPP columns
		if _, ok := m.definitionInput.ColumnsLPPSet[colName]; !ok {
			continue
		}

		newCol := m.Wiop.Columns.GetHandle(colName)

		if newCol.Round() != 0 {
			utils.Panic("expected a column with round 0, got %v, column: %v", newCol.Round(), colName)
		}

		colWitness, ok := witness.Columns[colName]
		if !ok {
			utils.Panic("witness of column %v was not found", colName)
		}

		run.AssignColumn(colName, colWitness)
		delete(witness.Columns, colName)
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

	moduleWitness := run.State.MustGet(moduleWitnessKey).(*ModuleWitness)
	run.State.Del(moduleWitnessKey)

	hornerParams := a.getHornerParams(run, moduleWitness.N0Values)

	run.AssignHornerParams(a.Horner.ID, hornerParams)
	run.AssignGrandProduct(a.GrandProduct.ID, a.GrandProduct.Compute(run))
	run.AssignLogDerivSum(a.LogDerivativeSum.ID, a.LogDerivativeSum.Compute(run))

	n0Hash, n1Hash := hashNxs(hornerParams, 0), hashNxs(hornerParams, 1)

	run.AssignColumn(a.N0Hash.GetColID(), smartvectors.NewRegular([]field.Element{n0Hash}))
	run.AssignColumn(a.N1Hash.GetColID(), smartvectors.NewRegular([]field.Element{n1Hash}))
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
		n0Hash        = hashNxsGnark(api, hornerParams, 0)
		n1Hash        = hashNxsGnark(api, hornerParams, 1)
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

	res := field.Element{}

	for _, part := range params.Parts {

		nx := 0

		if x == 0 {
			nx = part.N0
		} else {
			nx = part.N1
		}

		nxField := field.NewElement(uint64(nx))
		res = mimc.BlockCompression(res, nxField)
	}

	return res
}

// hashNxsGnark is as [hashNxs] but in a gnark circuit
func hashNxsGnark(api frontend.API, params query.GnarkHornerParams, x int) frontend.Variable {

	res := frontend.Variable(0)

	for _, part := range params.Parts {

		var nx frontend.Variable

		if x == 0 {
			nx = part.N0
		} else {
			nx = part.N1
		}

		res = mimc.GnarkBlockCompression(api, res, nx)
	}

	return res
}
