package tuning_test

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"testing"

	"github.com/consensys/linea-monorepo/prover/crypto/ringsis"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/fullrecursion"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/mimc"
	"github.com/consensys/linea-monorepo/prover/protocol/compiler/vortex"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/symbolic"
	"github.com/sirupsen/logrus"
)

func BenchmarkSelfRecursion(b *testing.B) {

	logrus.SetLevel(logrus.FatalLevel)

	var (
		numPoly = []int{40, 40, 40, 40}
		polSize = 1 << 18
	)

	defineFunc := func(bui *wizard.Builder) {

		for round := range numPoly {
			for i := 0; i < numPoly[round]; i++ {

				var (
					id = fmt.Sprintf("%v", rand.Int())
					c  = bui.RegisterCommit(ifaces.ColID("col"+id), polSize)
					_  = bui.GlobalConstraint(ifaces.QueryID("q"+id), symbolic.Mul(c, c))
				)
			}

			_ = bui.RegisterRandomCoin(coin.Name("coin"+strconv.Itoa(round)), coin.Field)
		}
	}

	sisParams := []ringsis.Params{
		// {LogTwoBound: 4, LogTwoDegree: 3},
		{LogTwoBound: 8, LogTwoDegree: 4},
		// {LogTwoBound: 16, LogTwoDegree: 6},
	}

	splitSizes := []int{
		// 1 << 12,
		// 1 << 13,
		// 1 << 14,
		1 << 15,
		// 1 << 16,
	}

	rsParams := []struct {
		BlowUp       int
		NumOpenedCol int
	}{
		// {BlowUp: 2, NumOpenedCol: 256},
		// {BlowUp: 4, NumOpenedCol: 128},
		{BlowUp: 16, NumOpenedCol: 64},
		// {BlowUp: 2, NumOpenedCol: 128},
		// {BlowUp: 4, NumOpenedCol: 64},
		// {BlowUp: 16, NumOpenedCol: 32},
		// {BlowUp: 256, NumOpenedCol: 16},
	}

	for _, sis := range sisParams {
		for _, split := range splitSizes {
			for _, rs := range rsParams {

				name := fmt.Sprintf(
					"sis.bound=%v-sis.deg=%v-split=%v-rs.blowup=%v-rs.numcolopened=%v",
					sis.LogTwoBound, sis.LogTwoDegree, split, rs.BlowUp, rs.NumOpenedCol,
				)

				b.Run(name, func(b *testing.B) {

					comp := wizard.Compile(
						defineFunc,
						compiler.Arcane(1<<8, split, true),
						vortex.Compile(rs.BlowUp, vortex.ForceNumOpenedColumns(rs.NumOpenedCol), vortex.WithSISParams(&sis)),
						fullrecursion.FullRecursion(true),
						mimc.CompileMiMC,
						compiler.Arcane(1<<8, split, true),
						vortex.Compile(rs.BlowUp, vortex.ForceNumOpenedColumns(rs.NumOpenedCol), vortex.WithSISParams(&sis)),
						fullrecursion.FullRecursion(true),
						mimc.CompileMiMC,
						compiler.Arcane(1<<8, split, true),
					)

					var (
						totalCells   = 0
						allCommitted = comp.Columns.AllKeysCommitted()
					)

					for _, name := range allCommitted {
						totalCells += comp.Columns.GetSize(name)
					}

					b.ReportMetric(float64(totalCells), "committed-cells-at-the-end")
				})
			}
		}
	}
}
