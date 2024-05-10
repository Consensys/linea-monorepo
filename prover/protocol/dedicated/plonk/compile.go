package plonk

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/iop"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/zkevm-monorepo/prover/maths/field"
	"github.com/consensys/zkevm-monorepo/prover/protocol/accessors"
	"github.com/consensys/zkevm-monorepo/prover/protocol/coin"
	"github.com/consensys/zkevm-monorepo/prover/protocol/dedicated/bigrange"
	"github.com/consensys/zkevm-monorepo/prover/protocol/dedicated/expr_handle"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/variables"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils"
)

// PlonkCheck adds a PLONK circuit in the wizard. Namely, the function takes a
// frontend.Circuit parameter, a PLONK witness assigner (i.e. a function that
// returns the PLONK witness to be used as an input for the solver). It
// compiles the circuit and construct a set of column and constraints reflecting
// the satisfiability of the provided PLONK circuit within the current Wizard.
// This is used for the precompiles and for ECDSA verification since these
// use-cases would require a very complex design if we wanted to implement them
// directly into the Wizard instead of Plonk.
//
// The user can provide one or more assigner for the same circuit to mean that
// we want to call the same circuit multiple times. In this case, the function
// optimizes the generated Wizard to commit only once to the preprocessed
// polynomials (qL, qR, etc...). This additionally allows batching certain parts
// of the protocol such as the copy-constraints argument which will be run only
// once over a random linear combination of the witnesses.
//
// The user can provide an identifying string `name` to the function. The name
// will be appended to all generated columns and queries name to carry some
// context to where these queries and columns come from.
func PlonkCheck(
	comp *wizard.CompiledIOP,
	name string,
	round int,
	circuit frontend.Circuit,
	// function to call to get an assignment
	witnessAssigner []func() frontend.Circuit,
	options ...Option,
) {
	// Create the ctx
	ctx := createCtx(comp, name, round, circuit, witnessAssigner, options...)

	// And registers the columns + constraints
	ctx.commitGateColumns()
	ctx.extractPermutationColumns()
	ctx.addCopyConstraint()
	ctx.addGateConstraint()

	if ctx.HasCommitment() {
		ctx.RegisterCommitProver()
	} else {
		ctx.registerNoCommitProver()
	}

	if ctx.RangeCheck.Enabled {
		ctx.addRangeCheckConstraint()
	}
}

// This function registers the Plonk gate's columns inside of the wizard. It
// does not add any constraints whatsoever.
func (ctx *compilationCtx) commitGateColumns() {

	// Declare and pre-assign the selector columns
	ctx.Columns.Ql = ctx.comp.InsertPrecomputed(ctx.colIDf("QL"), iopToSV(ctx.Plonk.Trace.Ql))
	ctx.Columns.Qr = ctx.comp.InsertPrecomputed(ctx.colIDf("QR"), iopToSV(ctx.Plonk.Trace.Qr))
	ctx.Columns.Qo = ctx.comp.InsertPrecomputed(ctx.colIDf("QO"), iopToSV(ctx.Plonk.Trace.Qo))
	ctx.Columns.Qm = ctx.comp.InsertPrecomputed(ctx.colIDf("QM"), iopToSV(ctx.Plonk.Trace.Qm))
	ctx.Columns.Qk = ctx.comp.InsertPrecomputed(ctx.colIDf("QK"), iopToSV(ctx.Plonk.Trace.Qk))

	// Declare and pre-assign the rangecheck selectors
	PcRcL, PcRcR := ctx.rcGetterToSV()
	ctx.Columns.RcL = ctx.comp.InsertPrecomputed(ctx.colIDf("RcL"), PcRcL)
	ctx.Columns.RcR = ctx.comp.InsertPrecomputed(ctx.colIDf("RcR"), PcRcR)

	ctx.Columns.L = make([]ifaces.Column, ctx.nbInstances)
	ctx.Columns.R = make([]ifaces.Column, ctx.nbInstances)
	ctx.Columns.O = make([]ifaces.Column, ctx.nbInstances)
	ctx.Columns.PI = make([]ifaces.Column, ctx.nbInstances)
	ctx.Columns.Cp = make([]ifaces.Column, ctx.nbInstances)

	if ctx.HasCommitment() {
		// Selector for the commitment
		ctx.Columns.Qcp = ctx.comp.InsertPrecomputed(ctx.colIDf("QCP"), iopToSV(ctx.Plonk.Trace.Qcp[0]))

		// First round, for the committed value and the PI
		for i := 0; i < ctx.nbInstances; i++ {
			ctx.Columns.PI[i] = ctx.comp.InsertCommit(ctx.round, ctx.colIDf("PI_%v", i), ctx.DomainSize())
			ctx.Columns.Cp[i] = ctx.comp.InsertCommit(ctx.round, ctx.colIDf("Cp_%v", i), ctx.DomainSize())
		}

		// Second rounds, after sampling HCP
		ctx.Columns.Hcp = ctx.comp.InsertCoin(ctx.round+1, coin.Name(ctx.Sprintf("HCP")), coin.Field)

		// And assigns the LRO polynomials
		for i := 0; i < ctx.nbInstances; i++ {
			ctx.Columns.L[i] = ctx.comp.InsertCommit(ctx.round+1, ctx.colIDf("L_%v", i), ctx.DomainSize())
			ctx.Columns.R[i] = ctx.comp.InsertCommit(ctx.round+1, ctx.colIDf("R_%v", i), ctx.DomainSize())
			ctx.Columns.O[i] = ctx.comp.InsertCommit(ctx.round+1, ctx.colIDf("O_%v", i), ctx.DomainSize())
		}
	} else {
		// Else no additional selector, and just commit to LRO + PI at the same round
		for i := 0; i < ctx.nbInstances; i++ {
			ctx.Columns.PI[i] = ctx.comp.InsertCommit(ctx.round, ctx.colIDf("PI_%v", i), ctx.DomainSize())
			ctx.Columns.L[i] = ctx.comp.InsertCommit(ctx.round, ctx.colIDf("L_%v", i), ctx.DomainSize())
			ctx.Columns.R[i] = ctx.comp.InsertCommit(ctx.round, ctx.colIDf("R_%v", i), ctx.DomainSize())
			ctx.Columns.O[i] = ctx.comp.InsertCommit(ctx.round, ctx.colIDf("O_%v", i), ctx.DomainSize())
		}
	}
}

// Returns an smart-vector from an iop.Polynomial
func iopToSV(pol *iop.Polynomial) smartvectors.SmartVector {
	return smartvectors.NewRegular(pol.Coefficients())
}

// This function constructs the PcRcL and PcRcR vectors. It is used by the
// external range-checking mechanism. Namely, the two constructed vectors are
// binary vectors containing a 1 to indicate that a wire is to be range-checked
// and 0 to indicate that it can be ignored.
//
// For instance, if it PcRcL[56] == 1, it means that the 56-th position of the
// PLONK column xA needs to be range-checked. For PcRcR, it indicates the
// relevant positions in the xB PLONK column.
//
// This function works by calling the RcGetter (see
// [plonk.newExternalRangeChecker] to obtain the result in "[]field.Element"
// form and then it converts it into assignable smartvectors after having
// checked a few hypothesis.
func (ctx *compilationCtx) rcGetterToSV() (PcRcL, PcRcR smartvectors.SmartVector) {
	v := [2][]field.Element{
		make([]field.Element, ctx.DomainSize()),
		make([]field.Element, ctx.DomainSize()),
	}
	sls := ctx.Plonk.RcGetter()
	for _, ss := range sls {
		v[ss[1]][ss[0]].SetInt64(1)
	}

	// check that v[0] and v[1] are not one at the same time
	for i := range v[0] {
		if v[0][i].IsOne() && v[1][i].IsOne() {
			utils.Panic(
				"broken assumption: v[0][i] = %v and v[1][i] = %v",
				v[0][i].String(), v[1][i].String(),
			)
		}
	}

	PcRcL = smartvectors.NewRegular(v[0])
	PcRcR = smartvectors.NewRegular(v[1])
	return PcRcL, PcRcR
}

// Extract the permutation columns and track them in the ctx
func (ctx *compilationCtx) extractPermutationColumns() {
	for i := range ctx.Columns.S {
		// Directly use the ints from the trace instead of the fresh Plonk ones
		si := ctx.Plonk.Trace.S[i*ctx.DomainSize() : (i+1)*ctx.DomainSize()]
		sField := make([]field.Element, len(si))
		for j := range sField {
			sField[j].SetInt64(si[j])
		}

		// Track it, no need to register it since the compiler
		// will do it on its own.
		ctx.Columns.S[i] = smartvectors.NewRegular(sField)
	}
}

// add gate constraint
func (ctx *compilationCtx) addGateConstraint() {

	for i := 0; i < ctx.nbInstances; i++ {

		// Declare the expression
		exp := ifaces.ColumnAsVariable(ctx.Columns.L[i]).Mul(ifaces.ColumnAsVariable(ctx.Columns.Ql)).
			Add(ifaces.ColumnAsVariable(ctx.Columns.R[i]).Mul(ifaces.ColumnAsVariable(ctx.Columns.Qr))).
			Add(ifaces.ColumnAsVariable(ctx.Columns.O[i]).Mul(ifaces.ColumnAsVariable(ctx.Columns.Qo))).
			Add(ifaces.ColumnAsVariable(ctx.Columns.L[i]).Mul(ifaces.ColumnAsVariable(ctx.Columns.R[i])).Mul(ifaces.ColumnAsVariable(ctx.Columns.Qm))).
			Add(ifaces.ColumnAsVariable(ctx.Columns.Qk).Add(ifaces.ColumnAsVariable(ctx.Columns.PI[i])))

		roundLRO := ctx.round

		// Optionally add a commitment
		if ctx.HasCommitment() {
			// full length of a column
			fullLength := ctx.Columns.PI[i].Size()
			hcpPosition := ctx.CommitmentInfo().CommitmentIndex + ctx.Plonk.SPR.GetNbPublicVariables()
			exp = exp.Add(
				ifaces.ColumnAsVariable(ctx.Columns.Qcp).
					Mul(ifaces.ColumnAsVariable(ctx.Columns.Cp[i])),
			).Add(
				ctx.Columns.Hcp.AsVariable().Mul(
					// equivalent to using Lagrange
					variables.NewPeriodicSample(fullLength, hcpPosition),
				),
			)
			// increase the LRO
			roundLRO++
		}

		// And registers the gate expression as a global variable
		ctx.comp.InsertGlobal(roundLRO, ctx.queryIDf("GATE_CS_INSTANCE_%v", i), exp)
	}
}

// add add the copy constraint
func (ctx *compilationCtx) addCopyConstraint() {

	// Creates a special handle for the permutation by
	// computing a linear combination of the columns
	var l, r, o ifaces.Column
	roundPermutation := ctx.Columns.L[0].Round()

	if len(ctx.Columns.L) == 1 {
		// then just pass the first column
		l = ctx.Columns.L[0]
		r = ctx.Columns.R[0]
		o = ctx.Columns.O[0]
	} else {
		// other run the permutation only once
		// over a linear combination of the columns
		roundPermutation++
		// declare the coin
		randLin := ctx.comp.InsertCoin(
			roundPermutation,
			coin.Name(ctx.Sprintf("PERMUTATION_RANDLIN")),
			coin.Field,
		)
		// And declare special columns for the linear combination
		l = expr_handle.RandLinCombCol(
			ctx.comp,
			accessors.NewFromCoin(randLin),
			ctx.Columns.L,
			ctx.Sprintf("L_PERMUT_LINCOMB"),
		)
		r = expr_handle.RandLinCombCol(
			ctx.comp,
			accessors.NewFromCoin(randLin),
			ctx.Columns.R,
			ctx.Sprintf("R_PERMUT_LINCOMB"),
		)
		o = expr_handle.RandLinCombCol(
			ctx.comp,
			accessors.NewFromCoin(randLin),
			ctx.Columns.O,
			ctx.Sprintf("O_PERMUT_LINCOMB"),
		)

	}

	// No need to commit to the permutation S =(s1,s2,s3),
	// as it is commited by FixedPermutation
	ctx.comp.InsertFixedPermutation(
		roundPermutation,
		ctx.queryIDf("PLONK_COPY_CS"),
		ctx.Columns.S[:],
		[]ifaces.Column{l, r, o},
		[]ifaces.Column{l, r, o},
	)
}

func (ctx *compilationCtx) addRangeCheckConstraint() {
	rcL := ifaces.ColumnAsVariable(ctx.Columns.RcL)
	rcR := ifaces.ColumnAsVariable(ctx.Columns.RcR)

	for i := range ctx.Columns.L {

		l := ifaces.ColumnAsVariable(ctx.Columns.L[i])
		r := ifaces.ColumnAsVariable(ctx.Columns.R[i])
		selectedL := rcL.Mul(l)
		selectedR := rcR.Mul(r)
		selected := selectedL.Add(selectedR)

		bigrange.BigRange(ctx.comp, selected, ctx.RangeCheck.NbLimbs, ctx.RangeCheck.NbBits, fmt.Sprintf("%v_%v", ctx.name, i))
	}
}
