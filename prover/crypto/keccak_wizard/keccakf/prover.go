package keccak

import (
	"sync"

	"github.com/consensys/accelerated-crypto-monorepo/maths/field"
	"github.com/consensys/accelerated-crypto-monorepo/protocol/wizard"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/parallel"
)

// the prover start assigning the values to the the columns (row by row)
func (ctx *KeccakFModule) ProverAssign(run *wizard.ProverRuntime) {
	// it sets the witnesses to the right length (length SIZE)
	ctx.BuildColumns()
	parallel.Execute(ctx.NP, func(start, stop int) {
		for n := start; n < stop; n++ {
			output := ctx.KeccakF1600(ctx.InputPI[n], n)
			if output != ctx.OutputPI[n] {
				utils.Panic(" the output of permutation is not correct for %v-th permutation", n)
			}

		}
	})
	// it assign commitment to the witnesses
	ctx.CommitColumns(run)
}

func (ctx *KeccakFModule) KeccakF1600(input [5][5]field.Element, n int) [5][5]field.Element {
	m := n * P2nRound
	w := ctx.witness
	// it assigns row by row
	w.a[m] = input
	for l := m; l < m+P2nRound; l++ {
		//aThetaFirst 65 bits
		w.aTheta[l] = ctx.assignAThetaFirst(w.a[l])
		// it returns the slices in both base-first and bit-base-second
		w.aTheta64[l], w.aThetaFirstSlice[l], w.aThetaSecondSlice[l], w.aTargetSliceDecompose[l], w.msb[l] = ctx.assignAThetaSlice(w.aTheta[l])
		// it assigns aRho in bit-base-second
		w.aRho[l] = ctx.assignARho(w.aThetaSecondSlice[l], w.aTargetSliceDecompose[l])
		//it assigns aPi in bit-base-second
		w.aPi[l] = ctx.assignAPi(w.aRho[l])
		//it finds the round constant for this round
		w.rcFieldSecond[l] = ctx.assignRCfieldSecond(RC, l%P2nRound)
		// assigns aChi
		w.aChiSecond[l] = ctx.assignAChiArith(w.aPi[l], w.rcFieldSecond[l])
		// it moves slice by slice from aChi to aIota; from base-second to bit-base-first
		w.aChiSecondSlice[l], w.aChiFirstSlice[l], w.aChiFirst[l] = ctx.assignAIotaChi(w.aChiSecond[l])
		if l < m+(P2nRound-1) {
			w.a[l+1] = w.aChiFirst[l]
		}
	}
	ctx.witness = w
	output := w.aChiFirst[m+(nRound-1)]
	return output
}
func (ctx *KeccakFModule) CommitColumns(run *wizard.ProverRuntime) {
	w := ctx.witness
	cn := ctx.colName
	// it builds the columns  and assign commitment

	// schedule tasks to run them in concurrently
	wg := &sync.WaitGroup{}
	schedule := func(f func()) {
		wg.Add(1)
		go func() {
			f()
			wg.Done()
		}()
	}

	schedule(func() { ctx.assignCommitColumnState(run, w.a, cn.NameA) })
	schedule(func() { ctx.assignCommitColumnState(run, w.aTheta, cn.NameAThetaFirst) })
	schedule(func() { ctx.assignCommitColumnSlice(run, w.aThetaFirstSlice, cn.NameAThetaFirstSlice) })
	schedule(func() { ctx.assignCommitColumnState(run, w.msb, cn.NamemsbAThetaFirst) })
	schedule(func() { ctx.assignCommitColumnSlice(run, w.aThetaSecondSlice, cn.NameAThetaSecondSlice) })
	schedule(func() { ctx.assignCommitColumnDecompose(run, w.aTargetSliceDecompose, cn.NameTargetSliceDecompos) })

	schedule(func() { ctx.assignCommitColumnState(run, w.aRho, cn.NameARho) })
	schedule(func() { ctx.assignCommitColumnState(run, w.aChiSecond, cn.NameAChiArith) })
	schedule(func() { ctx.assignCommitColumnRC(run, w.rcFieldSecond, cn.NameRCcolumn) })

	schedule(func() { ctx.assignCommitColumnSlice(run, w.aChiSecondSlice, cn.NameAChiArithSlice) })
	schedule(func() { ctx.assignCommitColumnSlice(run, w.aChiFirstSlice, cn.NameAIotaChiSlice) })

	//building and Assigning commitment to the Tables (used for changing the bases)
	schedule(func() { ctx.BuildTableFirstToSecond(run) })
	schedule(func() { ctx.BuildTableSecondToFirst(run) })

	wg.Wait()

}
