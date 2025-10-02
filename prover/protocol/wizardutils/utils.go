package wizardutils

import (
	"fmt"
	"strings"

	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/variables"
	"github.com/consensys/linea-monorepo/prover/protocol/zk"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// Parse the metadata of an expression and returns the first round where this function can be evaluated
func LastRoundToEval[T zk.Element](expr *symbolic.Expression[T]) int {
	board := expr.Board()
	metadatas := board.ListVariableMetadata()

	maxRound := 0

	for _, m := range metadatas {
		switch metadata := m.(type) {
		case ifaces.Column[T]:
			maxRound = utils.Max(maxRound, metadata.Round())
		// The expression can involve random coins
		case coin.Info[T]:
			maxRound = utils.Max(maxRound, metadata.Round)
			// assert the coin is an expression
			if metadata.Type == coin.IntegerVec {
				utils.Panic("The coin %v should be of type `Field`", metadata.Name)
			}
		case variables.X[T], variables.PeriodicSample[T]:
			// Do nothing
		case ifaces.Accessor[T]:
			maxRound = utils.Max(maxRound, metadata.Round())
		default:
			panic("unreachable")
		}
	}

	return maxRound
}

// DeriveName is used to construct either [ifaces.QueryID] or [ifaces.ColID] or
// [coin.Name]. The function will format [ifaces.Query], [ifaces.Column] or
// [coin.Info] using their names or IDs, in the other cases it will use the
// default `%v` formatting. The final returned name is obtained by stitching
// all the formatted inputs with an underscore.
func DeriveName[R ~string, T zk.Element](ss ...any) R {

	fmtted := []string{}

	for i := range ss {
		switch s := ss[i].(type) {
		case ifaces.Column[T]:
			fmtted = append(fmtted, string(s.GetColID()))
		case ifaces.Query[T]:
			fmtted = append(fmtted, string(s.Name()))
		case coin.Info[T]:
			fmtted = append(fmtted, string(s.Name))
		default:
			st := fmt.Sprintf("%v", s)
			fmtted = append(fmtted, st)
		}
	}

	return R(strings.Join(fmtted, "_"))
}

// AsExpr convert x into an expression expecting x to be either an expression and
// a column. The function also ensure that the expression can safely be
// interpreted as a column.
func AsExpr[T zk.Element](x any) (e *symbolic.Expression[T], round, size int) {

	switch c1 := x.(type) {
	case ifaces.Column[T]:
		round = c1.Round()
		size = c1.Size()
		e = symbolic.NewVariable[T](c1)
		return e, round, size
	case *symbolic.Expression[T]:
		board := c1.Board()
		e = c1
		size = column.ExprIsOnSameLengthHandles(&board)
		round = LastRoundToEval(c1)
		return e, round, size
	}

	panic("unexpected type")
}
