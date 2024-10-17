package wizard

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/zkevm-monorepo/prover/maths/common/smartvectors"
)

type Runtime interface {
	tryGetColumn(col *ColNatural) (smartvectors.SmartVector, bool)
	tryGetQueryRes(q Query) (QueryResult, bool)
	tryGetCoin(c Coin) (any, bool)
	getOrComputeQueryRes(q Query) QueryResult
}

type RuntimeGnark interface {
	tryGetColumn(col *ColNatural) ([]frontend.Variable, bool)
	tryGetQueryRes(q Query) (QueryResultGnark, bool)
	tryGetCoin(c Coin) (any, bool)
}
