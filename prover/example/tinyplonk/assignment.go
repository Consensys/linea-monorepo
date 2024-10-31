package tinyplonk

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/fft"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/iop"
	plonkBLS12_377 "github.com/consensys/gnark/backend/plonk/bls12-377"
	cs "github.com/consensys/gnark/constraint/bls12-377"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// DefineFromGnark constructs a [TinyPlonk] taking as inputs a [frontend.Circuit].
// It is more friendly to use than writting the constraints by hands and calling
// [DefineRaw].
func DefineFromGnark(comp *wizard.CompiledIOP, circ frontend.Circuit) *TinyPlonkCS {

	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), scs.NewBuilder, circ)
	if err != nil {
		utils.Panic("error compiling the circuit with name : %v", err)
	}

	var (
		spr        = ccs.(*cs.SparseR1CS)
		numPublic  = len(spr.Public)
		domainSize = utils.NextPowerOfTwo(spr.NbConstraints + len(spr.Public))
		domain     = fft.NewDomain(uint64(domainSize))
		trace      = plonkBLS12_377.NewTrace(spr, domain)
		sa         = make(sv.Regular, domainSize)
		sb         = make(sv.Regular, domainSize)
		sc         = make(sv.Regular, domainSize)
		ql         = iopToSV(trace.Ql)
		qr         = iopToSV(trace.Qr)
		qo         = iopToSV(trace.Qo)
		qm         = iopToSV(trace.Qm)
		qc         = iopToSV(trace.Qk)
	)

	for i := 0; i < domainSize; i++ {
		sa[i].SetInt64(trace.S[i])
		sb[i].SetInt64(trace.S[i+domainSize])
		sc[i].SetInt64(trace.S[i+2*domainSize])
	}

	plk := DefineRaw(comp, ql, qr, qo, qm, qc, &sa, &sb, &sc, numPublic)
	plk.spr = spr

	return plk
}

// AssignFromGnark constructs and assign a witness from a gnark circuits assignment.
// It is a wrapper around [Assign] that is simpler to use [Assign] as it performs
// the resolution of the assignment via the gnark solver and remove that burden
// from the caller.
//
// The receiver **must** have been constructed via the [DefineFromGnark] for this
// to work.
func (plk *TinyPlonkCS) AssignFromGnark(run *wizard.ProverRuntime, a frontend.Circuit) {

	if plk.spr == nil {
		utils.Panic("the [%T] must have been constructed via [DefineWithGnark]", plk)
	}

	witness, err := frontend.NewWitness(a, ecc.BLS12_377.ScalarField())
	if err != nil {
		utils.Panic("gnark could not parse the witness: %v", err.Error())
	}

	pubWitness, err := witness.Public()
	if err != nil {
		utils.Panic("public witness: %v", err.Error())
	}

	solution, err := plk.spr.Solve(witness)
	if err != nil {
		utils.Panic("Error in the solver: %v", err.Error())
	}

	sol := solution.(*cs.SparseR1CSSolution)

	run.AssignColumn(plk.Xa.GetColID(), sv.NewRegular(sol.L))
	run.AssignColumn(plk.Xb.GetColID(), sv.NewRegular(sol.R))
	run.AssignColumn(plk.Xc.GetColID(), sv.NewRegular(sol.O))
	run.AssignColumn(plk.Pi.GetColID(), sv.RightZeroPadded(pubWitness.Vector().(fr.Vector), len(sol.L)))

}

// Returns an smart-vector from an iop.Polynomial
func iopToSV(pol *iop.Polynomial) sv.SmartVector {
	return sv.NewRegular(pol.Coefficients())
}
