package ifaces

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext/gnarkfext"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
)

// Runtime is implemented by the [github.com/consensys/linea-monorepo/protocol/wizard.ProverRuntime] and
// [wizard.VerifierRuntime]. The interface exists to prevent circular
// dependencies internally.
type Runtime interface {
	// GetColumn returns the assignment of a registered column.
	GetColumn(ColID) ColAssignment
	// GetColumnAt returns a particular position of an assigned column.
	GetColumnAt(ColID, int) field.Element
	GetColumnAtBase(ColID, int) (field.Element, error)
	GetColumnAtExt(ColID, int) fext.Element
	// GetRandomCoinField returns the value of a random challenge coin
	GetRandomCoinField(name coin.Name) field.Element
	GetRandomCoinFieldExt(name coin.Name) fext.Element
	// GetRandomCoinIntegerVec returns the value of a coin.IntegerVec coin
	GetRandomCoinIntegerVec(name coin.Name) []int
	// GetParams returns the runtime parameters of a query
	GetParams(id QueryID) QueryParams
}

// Interface implemented by the [github.com/consensys/linea-monorepo/protocol/wizard.WizardVerifierCircuit]. The interface
// exists to prevent circular dependencies internally.
type GnarkRuntime interface {
	// GetColumn is as [Runtime.GetColumn] but in a gnark circuit
	GetColumn(ColID) []frontend.Variable
	GetColumnBase(ColID) ([]frontend.Variable, error)
	GetColumnExt(ColID) []gnarkfext.Element
	// GetColumnAt is as [Runtime.GetColumnAt] but in a gnark circuit
	GetColumnAt(ColID, int) frontend.Variable
	GetColumnAtBase(ColID, int) (frontend.Variable, error)
	GetColumnAtExt(ColID, int) gnarkfext.Element
	// GetRandomCoinField is as [Runtime.GetRandomCoinField] but in a gnark circuit
	GetRandomCoinField(name coin.Name) frontend.Variable
	GetRandomCoinFieldExt(name coin.Name) gnarkfext.Element
	// GetRandomCoinIntegerVec is as [Runtime.GetRandomCoinIntegerVec] but in a gnark circuit
	GetRandomCoinIntegerVec(name coin.Name) []frontend.Variable
	// GetParams is as [Runtime.GetParams] but in a gnark circuit
	GetParams(id QueryID) GnarkQueryParams
}
