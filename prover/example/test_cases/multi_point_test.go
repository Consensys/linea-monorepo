//go:build !fuzzlight

package test_cases_test

import (
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/protocol/coin"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
)

func defineMultiPoint(build *wizard.Builder) {
	P1 := build.RegisterCommit(P1, 4)
	build.RegisterRandomCoin(COIN1, coin.Field)
	P2 := build.RegisterCommit(P2, 4)
	build.RegisterRandomCoin(COIN2, coin.Field)
	build.UnivariateEval(UNIV1, P1)
	build.UnivariateEval(UNIV2, P2)
}

func proverProveMultiPoint(run *wizard.ProverRuntime) {
	/*
		It is not visible in the define but there is a constraint that
		forces p2 = 2 p1
	*/
	p1 := smartvectors.ForTest(1, 2, 3, 4)

	run.AssignColumn(P1, p1)
	c1 := run.GetRandomCoinField(COIN1)

	p2 := smartvectors.ForTest(2, 4, 6, 8)

	run.AssignColumn(P2, p2)
	c2 := run.GetRandomCoinField(COIN2)

	y1 := smartvectors.Interpolate(p1, c1)
	y2 := smartvectors.Interpolate(p2, c2)

	run.AssignUnivariate(UNIV1, c1, y1)
	run.AssignUnivariate(UNIV2, c2, y2)
}

func TestMultiPoint(t *testing.T) {
	checkSolved(t, defineMultiPoint, proverProveMultiPoint, UNIVARIATES, true)
}
