package query

import (
	"fmt"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/common/poly"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/utils"
)

type ProjectionInput struct {
	ColumnA, ColumnB []ifaces.Column
	FilterA, FilterB ifaces.Column
}
type Projection struct {
	Round int
	ID    ifaces.QueryID
	Inp   ProjectionInput
}

// NewProjection constructs a projection. Will panic if it is mal-formed
func NewProjection(
	round int,
	id ifaces.QueryID,
	inp ProjectionInput,
) Projection {
	var (
		sizeA  = inp.FilterA.Size()
		sizeB  = inp.FilterB.Size()
		numCol = len(inp.ColumnA)
	)

	if len(inp.ColumnB) != numCol {
		utils.Panic("A and B must have the same number of columns")
	}

	if ifaces.AssertSameLength(inp.ColumnA...) != sizeA {
		utils.Panic("A and its filter do not have the same column sizes")
	}

	if ifaces.AssertSameLength(inp.ColumnB...) != sizeB {
		utils.Panic("B and its filter do not have the same column sizes")
	}
	return Projection{Round: round, ID: id, Inp: inp}
}

// Name implements the [ifaces.Query] interface
func (p Projection) Name() ifaces.QueryID {
	return p.ID
}

// Check implements the [ifaces.Query] interface
func (p Projection) Check(run ifaces.Runtime) error {
	var (
		numCols               = len(p.Inp.ColumnA)
		sizeA                 = p.Inp.ColumnA[0].Size()
		sizeB                 = p.Inp.ColumnB[0].Size()
		linCombRand, evalRand field.Element
		a                     = make([]ifaces.ColAssignment, numCols)
		b                     = make([]ifaces.ColAssignment, numCols)
		fA                    = p.Inp.FilterA.GetColAssignment(run).IntoRegVecSaveAlloc()
		fB                    = p.Inp.FilterB.GetColAssignment(run).IntoRegVecSaveAlloc()
		aLinComb              = make([]field.Element, sizeA)
		bLinComb              = make([]field.Element, sizeB)
	)
	_, errAlpha := linCombRand.SetRandom()
	_, errBeta := evalRand.SetRandom()
	if errAlpha != nil {
		// Cannot happen unless the entropy was exhausted
		panic(errAlpha)
	}
	if errBeta != nil {
		// Cannot happen unless the entropy was exhausted
		panic(errBeta)
	}
	// Populate a
	for colIndex, pol := range p.Inp.ColumnA {
		a[colIndex] = pol.GetColAssignment(run)
	}
	// Populate b
	for colIndex, pol := range p.Inp.ColumnB {
		b[colIndex] = pol.GetColAssignment(run)
	}
	// Compute the linear combination of the columns of a and b
	for row := 0; row < sizeA; row++ {
		aLinComb[row] = rowLinComb(linCombRand, row, a)
	}
	for row := 0; row < sizeB; row++ {
		bLinComb[row] = rowLinComb(linCombRand, row, b)
	}
	var (
		hornerA = poly.GetHornerTrace(aLinComb, fA, evalRand)
		hornerB = poly.GetHornerTrace(bLinComb, fB, evalRand)
	)
	if hornerA[0] != hornerB[0] {
		return fmt.Errorf("the projection query %v check is not satisfied", p.ID)
	}

	return nil
}

// GnarkCheck implements the [ifaces.Query] interface. It will panic in this
// construction because we do not have a good way to check the query within a
// circuit
func (i Projection) CheckGnark(api frontend.API, run ifaces.GnarkRuntime) {
	panic("UNSUPPORTED : can't check an Projection query directly into the circuit")
}

// GetShiftedRelatedColumns returns the list of the [HornerParts.Selectors] found
// in the query. This is used to check if the query is compatible with
// Wizard distribution.
//
// Note: the fact that this method is implemented makes [Inclusion] satisfy
// an anonymous interface that is matched to detect queries that are
// incompatible with wizard distribution. So we should not rename or remove
// this implementation without doing the corresponding changes in the
// distributed package. Otherwise, this will silence the checks that we are
// doing.
func (p Projection) GetShiftedRelatedColumns() []ifaces.Column {

	res := []ifaces.Column{}

	if p.Inp.FilterA.IsComposite() {
		res = append(res, p.Inp.FilterA)
	}

	if p.Inp.FilterB.IsComposite() {
		res = append(res, p.Inp.FilterB)
	}

	for i := range p.Inp.ColumnA {
		if p.Inp.ColumnA[i].IsComposite() {
			res = append(res, p.Inp.ColumnA[i])
		}
	}

	for i := range p.Inp.ColumnB {
		if p.Inp.ColumnB[i].IsComposite() {
			res = append(res, p.Inp.ColumnB[i])
		}
	}

	return res
}
