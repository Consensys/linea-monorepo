package accessors

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/coin"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/ifaces"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/query"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/gnark/frontend"
)

func AccessorFromCoin(info coin.Info) *ifaces.Accessor {

	if info.Type != coin.Field {
		panic("only supports coins")
	}

	return ifaces.NewAccessor(
		fmt.Sprintf("FROM_COIN_%v", info.Name),
		func(run ifaces.Runtime) field.Element { return run.GetRandomCoinField(info.Name) },
		func(api frontend.API, run ifaces.GnarkRuntime) frontend.Variable {
			return run.GetRandomCoinField(info.Name)
		},
		info.Round,
	)
}

func AccessorFromConstant(x field.Element) *ifaces.Accessor {
	return ifaces.NewAccessor(
		fmt.Sprintf("FROM_CONST_%v", x.String()),
		func(run ifaces.Runtime) field.Element { return x },
		func(api frontend.API, run ifaces.GnarkRuntime) frontend.Variable { return frontend.Variable(x) },
		0, // It's a constant so it's free
	)
}

func AccessorFromUnivX(comp *wizard.CompiledIOP, q query.UnivariateEval) *ifaces.Accessor {
	return ifaces.NewAccessor(
		fmt.Sprintf("FROM_UNIVX_%v", q.QueryID),
		func(run ifaces.Runtime) field.Element {
			return run.GetParams(q.QueryID).(query.UnivariateEvalParams).X
		},
		func(api frontend.API, run ifaces.GnarkRuntime) frontend.Variable {
			return run.GetParams(q.QueryID).(query.GnarkUnivariateEvalParams).X
		},
		comp.QueriesParams.Round(q.QueryID),
	)
}
