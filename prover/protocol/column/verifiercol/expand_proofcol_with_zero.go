package verifiercol

import (
    "github.com/consensys/gnark/frontend"
    "github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
    "github.com/consensys/linea-monorepo/prover/maths/common/vector"
    "github.com/consensys/linea-monorepo/prover/maths/field"
    "github.com/consensys/linea-monorepo/prover/protocol/ifaces"
    "github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

// compile check to enforce the struct to belong to the corresponding interface
var _ VerifierCol = ExpandedProofColWithZero{}

type ExpandedProofColWithZero struct {
    Proofcol  ifaces.Column
    Expansion int
}

// Round returns the round ID of the column and implements the [ifaces.Column]
// interface.
func (ex ExpandedProofColWithZero) Round() int {
    return ex.Proofcol.Round()
}

// GetColID returns the column ID
func (ex ExpandedProofColWithZero) GetColID() ifaces.ColID {
    return ifaces.ColIDf("EXPANDED_%v_%v_WITH_ZERO", ex.Proofcol.GetColID(), ex.Expansion)
}

// MustExists implements the [ifaces.Column] interface and always returns true.
func (ex ExpandedProofColWithZero) MustExists() {
    ex.Proofcol.MustExists()
}

// Size returns the size of the colum and implements the [ifaces.Column]
// interface.
func (ex ExpandedProofColWithZero) Size() int {
    return ex.Proofcol.Size() * ex.Expansion
}

// GetColAssignment returns the assignment of the current column
// Todo: Base columns are proof columns
func (ex ExpandedProofColWithZero) GetColAssignment(run ifaces.Runtime) ifaces.ColAssignment {
    assi := ex.Proofcol.GetColAssignment(run)
    values := make([][]field.Element, ex.Expansion)
    values[0] = smartvectors.IntoRegVec(assi)
    // zeros denote slice of zeros of length of the verifier column
    zeros := make([]field.Element, ex.Proofcol.Size())
    // we start at 1 because the first slice is the actual verifier column
    // the other slices are zeros
    for j := 1; j < ex.Expansion; j++ {
        values[j] = zeros
    }
    res := vector.Interleave(values...)
    return smartvectors.NewRegular(res)
}

// GetColAssignment returns a gnark assignment of the current column
func (ex ExpandedProofColWithZero) GetColAssignmentGnark(run ifaces.GnarkRuntime) []frontend.Variable {
    assi := ex.Proofcol.GetColAssignmentGnark(run)
    res := make([]frontend.Variable, ex.Size())
    for i := 0; i < len(assi); i++ {
        res[i*ex.Expansion] = assi[i]
        for j := 1; j < ex.Expansion; j++ {
            res[j+i*ex.Expansion] = frontend.Variable(0)
        }
    }
    return res
}

// GetColAssignmentAt returns a particular position of the column
func (ex ExpandedProofColWithZero) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
    return ex.Proofcol.GetColAssignmentAt(run, pos/ex.Expansion)
}

// GetColAssignmentGnarkAt returns a particular position of the column in a gnark circuit
func (ex ExpandedProofColWithZero) GetColAssignmentGnarkAt(run ifaces.GnarkRuntime, pos int) frontend.Variable {
    return ex.Proofcol.GetColAssignmentGnarkAt(run, pos/ex.Expansion)
}

// IsComposite implements the [ifaces.Column] interface
func (ex ExpandedProofColWithZero) IsComposite() bool {
    return ex.Proofcol.IsComposite()
}

// String implements the [symbolic.Metadata] interface
func (ex ExpandedProofColWithZero) String() string {
    return string(ex.GetColID())
}

// Split implements the [VerifierCol] interface,
// it is no-op for proofcols
func (ex ExpandedProofColWithZero) Split(_ *wizard.CompiledIOP, from, to int) ifaces.Column {
    return ex.Proofcol
}