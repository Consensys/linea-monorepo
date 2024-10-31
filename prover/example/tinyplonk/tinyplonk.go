package tinyplonk

import (
	"fmt"

	cs "github.com/consensys/gnark/constraint/bls12-377"
	"github.com/consensys/gnark/frontend"
	sv "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	sym "github.com/consensys/linea-monorepo/prover/symbolic"
)

// TinyPlonkCS is a wizard representation of a Plonk circuit.
type TinyPlonkCS struct {

	// Ql, Qr, Qo, Qm and Qc are precomputed columns describing the Plonk CS
	// we register.
	Ql, Qr, Qo, Qm, Qc ifaces.Column

	// Xa, Xb, Xc are the columns storing the witness of the CS
	Xa, Xb, Xc ifaces.Column

	// Pi collects the statement (the public inputs) of the CS. The verifier
	// has to manually check that the provided assignment to Pi consist in the
	// right number of values padded with 0s.
	Pi ifaces.Column

	// NbPublic indicates the number of public inputs of the circuit. This is used
	// by the verifier because this has an impact on the form of Pi: every
	// coordinate at position >= NbPublic must be zero.
	NbPublic int

	// spr is only populated if the object is constructed via the [DefineFromGnark]
	// function.
	spr *cs.SparseR1CS
}

// verifierDirectCheck checks the well-formedness of the PI
type verifierDirectCheck struct {
	TinyPlonkCS
}

// DefineRaw constructs and registers a [TinyPlonkCS] is in the current
// [wizard.CompiledIOP] object and returns the TinyPlonkCS. All provided vectors
// must have the same size and the size must be a power of two.
//
// The ql, qr, ..., qc are the values to provide to the "Q*" columns. And the
// sa, sb, sc are used to provide the permutation argument.
func DefineRaw(comp *wizard.CompiledIOP, ql, qr, qo, qm, qc, sa, sb, sc sv.SmartVector, nbPublic int) *TinyPlonkCS {

	nbConstraints := ql.Len()

	if ql.Len() != nbConstraints ||
		qr.Len() != nbConstraints ||
		qo.Len() != nbConstraints ||
		qm.Len() != nbConstraints ||
		qc.Len() != nbConstraints ||
		sa.Len() != nbConstraints ||
		sb.Len() != nbConstraints ||
		sc.Len() != nbConstraints {
		panic("inconsistent circuit description: all smartvector should have the same size")
	}

	// Implicitly, this check is preventing a circuit that only "loads" the
	// public inputs.
	if nbPublic >= nbConstraints {
		panic("more public inputs than there are constraints")
	}

	plk := &TinyPlonkCS{
		Ql:       comp.InsertPrecomputed("QL", ql),
		Qr:       comp.InsertPrecomputed("QR", qr),
		Qo:       comp.InsertPrecomputed("QO", qo),
		Qm:       comp.InsertPrecomputed("QM", qm),
		Qc:       comp.InsertPrecomputed("QC", qc),
		Xa:       comp.InsertCommit(0, "XA", nbConstraints),
		Xb:       comp.InsertCommit(0, "XB", nbConstraints),
		Xc:       comp.InsertCommit(0, "XC", nbConstraints),
		Pi:       comp.InsertProof(0, "PI", nbConstraints),
		NbPublic: nbPublic,
	}

	// This defines the gate constraint
	comp.InsertGlobal(
		0,
		"GATE-CS",
		sym.Add(
			sym.Mul(plk.Ql, plk.Xa),
			sym.Mul(plk.Qr, plk.Xb),
			sym.Mul(plk.Qm, plk.Xa, plk.Xb),
			sym.Mul(plk.Qo, plk.Xc),
			plk.Qc,
			plk.Pi,
		),
	)

	// This declares the copy constraints
	comp.InsertFixedPermutation(
		0,
		"COPY-CS",
		[]sv.SmartVector{sa, sb, sc},
		[]ifaces.Column{plk.Xa, plk.Xb, plk.Xc},
		[]ifaces.Column{plk.Xa, plk.Xb, plk.Xc},
	)

	// This tells the verifier to check the well-formedness of the PI vector
	comp.RegisterVerifierAction(0, &verifierDirectCheck{TinyPlonkCS: *plk})

	return plk
}

// Assign allows assigning the columns of the receiver [TinyPlonk] so that we
// construct the witness of a proof to generate. The caller is responsible to
// provide a solution to the Plonk circuit.
func (plk *TinyPlonkCS) Assign(run *wizard.ProverRuntime, xa, xb, xc, pi sv.SmartVector) {
	run.AssignColumn(plk.Xa.GetColID(), xa)
	run.AssignColumn(plk.Xb.GetColID(), xb)
	run.AssignColumn(plk.Xc.GetColID(), xc)
	run.AssignColumn(plk.Pi.GetColID(), pi)
}

// This implements the [wizard.VerifierAction] interface
func (v *verifierDirectCheck) Run(run *wizard.VerifierRuntime) error {

	pi := v.Pi.GetColAssignment(run)

	for i := v.NbPublic; i < pi.Len(); i++ {
		x := pi.Get(i)
		if !x.IsZero() {
			return fmt.Errorf("PI is not well-formed")
		}
	}

	return nil
}

// The function is only implemented for interface satisfaction matters. It is
// non-necessary as it is used for recursion that we don't do.
func (v *verifierDirectCheck) RunGnark(frontend.API, *wizard.WizardVerifierCircuit) {}
