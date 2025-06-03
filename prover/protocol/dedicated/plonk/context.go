package plonk

import (
	"sync"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	plonkBLS12_377 "github.com/consensys/gnark/backend/plonk/bls12-377"
	cs "github.com/consensys/gnark/constraint/bls12-377"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
	"github.com/sirupsen/logrus"
)

// The compilationCtx (context) carries all the compilation informations about a call to
// Plonk in Wizard. Namely, (non-exhaustively) it contains the gnark's internal
// informations about it, the generated Wizard columns and the compilation
// parameters that are used for the compilation. The context is also the
// receiver of all the methods allowing to construct the Plonk in Wizard module.
type compilationCtx struct {
	// The compiled IOP
	comp *wizard.CompiledIOP
	// Name of the context
	name string
	// Round at which we create the ctx
	round int
	// Number of instances of the circuit
	maxNbInstances int

	// Gnark related data
	Plonk struct {
		// The plonk circuit being integrated
		Circuit frontend.Circuit
		// The compiled circuit
		Trace *plonkBLS12_377.Trace
		// The sparse constrained system
		SPR *cs.SparseR1CS
		// Domain to gets the polynomials in lagrange form
		Domain *fft.Domain
		// Options for the solver, may contain hint informations
		// and so on.
		SolverOpts []solver.Option
		// Receives the list of rows which have to be marked containing range checks.
		RcGetter func() [][2]int // the same for all circuits
	}

	// Columns
	Columns struct {
		// Circuit columns
		Ql, Qr, Qm, Qo, Qk, Qcp ifaces.Column
		// Witness columns
		L, R, O, PI, TinyPI, Cp []ifaces.Column
		// Activators are tiny verifier-visible columns that are used to
		// deactivate the constraints happening for constraints that are not
		// happening in the system. The verifier is required to check that the
		// columns are assigned to binary values and that they are structured
		// as a sequence of 1s followed by a sequence of 0s.
		Activators []ifaces.Column
		// Columns representing the permutation
		S [3]ifaces.ColAssignment
		// Commitment randomness
		Hcp coin.Info
		// Selector for range checking from a column
		RcL, RcR, RcO ifaces.Column
		// RangeChecked stores the values to be range-checked
		RangeChecked []ifaces.Column
	}

	// Optional field used for specifying range checks
	RangeCheck struct {
		// wasCancelled is set if no wires need to be constrained
		wasCancelled         bool
		Enabled              bool
		NbBits               int
		NbLimbs              int
		AddGateForRangeCheck bool
		limbDecomposition    []wizard.ProverAction
	}
}

// Create the context from a circuit. Extracts all the
// PLONK data representing the circuit
func createCtx(
	comp *wizard.CompiledIOP,
	name string,
	round int,
	circuit frontend.Circuit,
	maxNbInstance int,
	opts ...Option,
) (ctx compilationCtx) {

	ctx = compilationCtx{
		comp:           comp,
		name:           name,
		round:          round,
		maxNbInstances: maxNbInstance,
	}

	ctx.Plonk.Circuit = circuit

	for _, opt := range opts {
		opt(&ctx)
	}

	// Build the trace and track it in the context
	gnarkBuilder, rcGetter := newExternalRangeChecker(comp, ctx.RangeCheck.AddGateForRangeCheck)

	logrus.Debugf("Plonk in Wizard (%v) compiling the circuit", name)
	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), gnarkBuilder, circuit)
	if err != nil {
		utils.Panic("error compiling the circuit with name `%v` : %v", name, err)
	}

	logrus.Debugf(
		"constraint system has %v constraints and %v internal variables",
		ccs.GetNbConstraints(), ccs.GetNbInternalVariables(),
	)

	ctx.Plonk.RcGetter = rcGetter // Pass the range-check getter
	ctx.Plonk.SPR = ccs.(*cs.SparseR1CS)
	ctx.Plonk.Domain = fft.NewDomain(uint64(ctx.DomainSize()))

	logrus.Debugf("Plonk in Wizard (%v) build trace", name)
	ctx.Plonk.Trace = plonkBLS12_377.NewTrace(ctx.Plonk.SPR, ctx.Plonk.Domain)

	logrus.Debugf("Plonk in Wizard (%v) build permutation", name)
	ctx.buildPermutation(ctx.Plonk.SPR, ctx.Plonk.Trace) // no part of BuildTrace

	logrus.Debugf("Plonk in Wizard (%v) done", name)
	return ctx
}

// Return the size of the domain
func (ctx *compilationCtx) DomainSize() int {
	// fft domains
	return utils.NextPowerOfTwo(
		ctx.Plonk.SPR.NbConstraints + len(ctx.Plonk.SPR.Public),
	)
}

// Returns the size of the public input tiny column
func (ctx *compilationCtx) TinyPISize() int {
	return utils.NextPowerOfTwo(
		ctx.Plonk.SPR.GetNbPublicVariables(),
	)
}

// Complete the Plonk.Trace by computing the S permutation.
// (Copied from gnark)
//
// buildPermutation builds the Permutation associated with a circuit.
//
// The permutation s is composed of cycles of maximum length such that
//
//	s. (l∥r∥o) = (l∥r∥o)
//
// , where l∥r∥o is the concatenation of the indices of l, r, o in
// ql.l+qr.r+qm.l.r+qo.O+k = 0.
//
// The permutation is encoded as a slice s of size 3*size(l), where the
// i-th entry of l∥r∥o is sent to the s[i]-th entry, so it acts on a tab
// like this: for i in tab: tab[i] = tab[permutation[i]]
func (ctx *compilationCtx) buildPermutation(spr *cs.SparseR1CS, pt *plonkBLS12_377.Trace) {

	nbVariables := spr.NbInternalVariables + len(spr.Public) + len(spr.Secret)

	// nbVariables := spr.NbInternalVariables + len(spr.Public) + len(spr.Secret)
	sizeSolution := ctx.DomainSize()
	sizePermutation := 3 * sizeSolution

	// init permutation
	permutation := make([]int64, sizePermutation)
	for i := 0; i < len(permutation); i++ {
		permutation[i] = -1
	}

	// init LRO position -> variable_ID
	lro := make([]int, sizePermutation) // position -> variable_ID
	for i := 0; i < len(spr.Public); i++ {
		lro[i] = i // IDs of LRO associated to placeholders (only L needs to be taken care of)
	}

	offset := len(spr.Public)

	j := 0
	it := spr.GetSparseR1CIterator()
	for c := it.Next(); c != nil; c = it.Next() {
		lro[offset+j] = int(c.XA)
		lro[sizeSolution+offset+j] = int(c.XB)
		lro[2*sizeSolution+offset+j] = int(c.XC)

		j++
	}

	// init cycle:
	// map ID -> last position the ID was seen
	cycle := make([]int64, nbVariables)
	for i := 0; i < len(cycle); i++ {
		cycle[i] = -1
	}

	for i := 0; i < len(lro); i++ {
		if cycle[lro[i]] != -1 {
			// if != -1, it means we already encountered this value
			// so we need to set the corresponding permutation index.
			permutation[i] = cycle[lro[i]]
		}
		cycle[lro[i]] = int64(i)
	}

	// complete the Permutation by filling the first IDs encountered
	for i := 0; i < sizePermutation; i++ {
		if permutation[i] == -1 {
			permutation[i] = cycle[lro[i]]
		}
	}

	pt.S = permutation
}

// ConcatenatedTinyPIs returns a verifier column such constructed by stacking
// all the `TinyPI` columns on top of one another and padding that with zeroes
// up to the desired size.
func (ctx compilationCtx) ConcatenatedTinyPIs(size int) ifaces.Column {
	return verifiercol.NewConcatTinyColumns(
		ctx.comp,
		size,
		field.Zero(),
		ctx.Columns.TinyPI...,
	)
}

// GetPlonkProverAction returns the [PlonkInWizardProverAction] responsible for
// assigning the first round of the wizard. In case we use the BBS commitment
// this stands for [initialBBSProverAction] or [noCommitProverAction] in the
// contrary case.
func (ctx compilationCtx) GetPlonkProverAction() PlonkInWizardProverAction {

	if ctx.HasCommitment() {
		return initialBBSProverAction{
			compilationCtx:  ctx,
			proverStateLock: &sync.Mutex{},
		}
	}

	return noCommitProverAction(ctx)
}
