package arithmetization

import (
	"reflect"
	"testing"

	"github.com/consensys/zkevm-monorepo/prover/config"
	"github.com/consensys/zkevm-monorepo/prover/protocol/coin"
	"github.com/consensys/zkevm-monorepo/prover/protocol/column"
	"github.com/consensys/zkevm-monorepo/prover/protocol/ifaces"
	"github.com/consensys/zkevm-monorepo/prover/protocol/wizard"
	"github.com/consensys/zkevm-monorepo/prover/utils/collection"
	"github.com/stretchr/testify/require"
)

func TestDefine(t *testing.T) {

	var (
		comp = &wizard.CompiledIOP{
			Columns:         column.NewStore(),
			QueriesParams:   wizard.NewRegister[ifaces.QueryID, ifaces.Query](),
			QueriesNoParams: wizard.NewRegister[ifaces.QueryID, ifaces.Query](),
			Coins:           wizard.NewRegister[coin.Name, coin.Info](),
			Precomputed:     collection.NewMapping[ifaces.ColID, ifaces.ColAssignment](),
		}
		schema, errBin = ReadZkevmBin()
		limits         = &config.TracesLimits{}
		limitRefl      = reflect.ValueOf(limits).Elem()
	)

	for i := 0; i < limitRefl.NumField(); i++ {
		limitRefl.Field(i).SetInt(1 << 10)
	}

	require.NoError(t, errBin)
	Define(comp, schema, limits)
}
