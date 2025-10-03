package verifiercol

import (
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vectorext"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
)

// compile check to enforce the struct to belong to the corresponding interface
// var _ VerifierCol = ExpandedVerifCol{}

type ExpandedVerifCol[T zk.Element] struct {
	Verifiercol VerifierCol[T]
	Expansion   int
}

func (ex ExpandedVerifCol[T]) IsBase() bool {
	return ex.Verifiercol.IsBase()
}

func (ex ExpandedVerifCol[T]) GetColAssignmentAtBase(run ifaces.Runtime, pos int) (field.Element, error) {
	//TODO implement me
	panic("implement me")
}

func (ex ExpandedVerifCol[T]) GetColAssignmentAtExt(run ifaces.Runtime, pos int) fext.Element {
	//TODO implement me
	panic("implement me")
}

func (ex ExpandedVerifCol[T]) GetColAssignmentGnarkBase(run ifaces.GnarkRuntime[T]) ([]T, error) {
	//TODO implement me
	panic("implement me")
}

func (ex ExpandedVerifCol[T]) GetColAssignmentGnarkExt(run ifaces.GnarkRuntime[T]) []gnarkfext.E4Gen[T] {
	//TODO implement me
	panic("implement me")
}

func (ex ExpandedVerifCol[T]) GetColAssignmentGnarkAtBase(run ifaces.GnarkRuntime[T], pos int) (T, error) {
	//TODO implement me
	panic("implement me")
}

func (ex ExpandedVerifCol[T]) GetColAssignmentGnarkAtExt(run ifaces.GnarkRuntime[T], pos int) gnarkfext.E4Gen[T] {
	//TODO implement me
	panic("implement me")
}

// Round returns the round ID of the column and implements the [ifaces.Column[T]]
// interface.
func (ex ExpandedVerifCol[T]) Round() int {
	return ex.Verifiercol.Round()
}

// GetColID returns the column ID
func (ex ExpandedVerifCol[T]) GetColID() ifaces.ColID {
	return ifaces.ColIDf("Expanded_%v_%v", ex.Verifiercol.GetColID(), ex.Expansion)
}

// MustExists implements the [ifaces.Column[T]] interface and always returns true.
func (ex ExpandedVerifCol[T]) MustExists() {
	ex.Verifiercol.MustExists()
}

// Size returns the size of the colum and implements the [ifaces.Column[T]]
// interface.
func (ex ExpandedVerifCol[T]) Size() int {
	return ex.Verifiercol.Size() * ex.Expansion
}

// GetColAssignment returns the assignment of the current column
func (ex ExpandedVerifCol[T]) GetColAssignment(run ifaces.Runtime) ifaces.ColAssignment {
	assi := ex.Verifiercol.GetColAssignment(run)
	values := make([][]fext.Element, ex.Expansion)
	for j := range values {
		values[j] = smartvectors.IntoRegVecExt(assi)
	}
	res := vectorext.Interleave(values...)
	return smartvectors.NewRegularExt(res)
}

// GetColAssignment returns a gnark assignment of the current column
func (ex ExpandedVerifCol[T]) GetColAssignmentGnark(run ifaces.GnarkRuntime[T]) []T {
	assi := ex.Verifiercol.GetColAssignmentGnark(run)
	res := make([]T, ex.Size())
	for i := 0; i < len(assi); i++ {
		for j := 0; j < ex.Expansion; j++ {
			res[j+i*ex.Expansion] = assi[i]
		}
	}
	return res
}

// GetColAssignmentAt returns a particular position of the column
func (ex ExpandedVerifCol[T]) GetColAssignmentAt(run ifaces.Runtime, pos int) field.Element {
	return ex.Verifiercol.GetColAssignmentAt(run, pos/ex.Expansion)
}

// GetColAssignmentGnarkAt returns a particular position of the column in a gnark circuit
func (ex ExpandedVerifCol[T]) GetColAssignmentGnarkAt(run ifaces.GnarkRuntime[T], pos int) T {
	return ex.Verifiercol.GetColAssignmentGnarkAt(run, pos/ex.Expansion)
}

// IsComposite implements the [ifaces.Column[T]] interface
func (ex ExpandedVerifCol[T]) IsComposite() bool {
	return ex.Verifiercol.IsComposite()
}

// String implements the [symbolic.Metadata] interface
func (ex ExpandedVerifCol[T]) String() string {
	return string(ex.GetColID())
}

// Split implements the [VerifierCol] interface
func (ex ExpandedVerifCol[T]) Split(_ *wizard.CompiledIOP[T], from, to int) ifaces.Column[T] {
	return ex.Verifiercol
}
