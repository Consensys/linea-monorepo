package plonkinternal

import (
	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/protocol/accessors"
	"github.com/consensys/linea-monorepo/prover/protocol/column/verifiercol"
	"github.com/consensys/linea-monorepo/prover/protocol/dedicated"
	"github.com/consensys/linea-monorepo/prover/protocol/ifaces"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/consensys/linea-monorepo/prover/utils"
)

// addHashConstraint adds the constraints the hashes of the tagged witness.
// It checks that the assignment of the hash column is consistent with the
// LRO columns using a lookup and then it uses a Poseidon2 query to enforce the
// hash.
func (ctx *CompilationCtx) addHashConstraint() {

	var (
		numRowLRO = ctx.DomainSize()
		round     = ctx.Columns.L[0].Round()
		eho       = &ctx.ExternalHasherOption

		// Records the positions of the hash claims in the Plonk rows.
		posOsSv, posBlSv, posNsSv = ctx.getHashCheckedPositionSV()
		size                      = posOsSv.Len()

		// Declare the L, R, O position columns. These will be cached and
		// reused by the fixed permutation compiler.
		posL = dedicated.CounterPrecomputed(ctx.comp, 0, numRowLRO)
		posR = dedicated.CounterPrecomputed(ctx.comp, numRowLRO, 2*numRowLRO)
		posO = dedicated.CounterPrecomputed(ctx.comp, 2*numRowLRO, 3*numRowLRO)
	)

	for j := 0; j < poseidon2_koalabear.BlockSize; j++ {
		eho.PosOldState[j] = ctx.comp.InsertPrecomputed(ctx.colIDf("HashCheckPositionOS_%v", j), posOsSv)
		eho.PosBlock[j] = ctx.comp.InsertPrecomputed(ctx.colIDf("HashCheckPositionBL_%v", j), posBlSv)
		eho.PosNewState[j] = ctx.comp.InsertPrecomputed(ctx.colIDf("HashCheckPositionNS_%v", j), posNsSv)
	}
	eho.OldStates = make([][poseidon2_koalabear.BlockSize]ifaces.Column, ctx.MaxNbInstances)
	eho.Blocks = make([][poseidon2_koalabear.BlockSize]ifaces.Column, ctx.MaxNbInstances)
	eho.NewStates = make([][poseidon2_koalabear.BlockSize]ifaces.Column, ctx.MaxNbInstances)

	for i := 0; i < ctx.MaxNbInstances; i++ {

		var (
			selector = verifiercol.NewRepeatedAccessor(
				accessors.NewFromPublicColumn(ctx.Columns.Activators[i], 0),
				size,
			)

			lookupTables = [][]ifaces.Column{
				{
					posL, ctx.Columns.L[i],
				},
				{
					posR, ctx.Columns.R[i],
				},
				{
					posO, ctx.Columns.O[i],
				},
			}
		)

		for j := 0; j < poseidon2_koalabear.BlockSize; j++ {
			eho.OldStates[i][j] = ctx.comp.InsertCommit(round, ctx.colIDf("HashCheckOldState_%v_%v", i, j), size, true)
			eho.Blocks[i][j] = ctx.comp.InsertCommit(round, ctx.colIDf("HashCheckBlock_%v_%v", i, j), size, true)
			eho.NewStates[i][j] = ctx.comp.InsertCommit(round, ctx.colIDf("HashCheckNewState_%v_%v", i, j), size, true)

			// Those are lookups checking that the LRO columns are consistent with
			// the hash claims.

			ctx.comp.GenericFragmentedConditionalInclusion(
				round,
				ctx.queryIDf("HashCheckImportOldState_%v_%v", i, j),
				// including
				lookupTables,
				// included
				[]ifaces.Column{
					eho.PosOldState[j], eho.OldStates[i][j],
				},
				// no filters
				nil, nil,
			)

			ctx.comp.GenericFragmentedConditionalInclusion(
				round,
				ctx.queryIDf("HashCheckImportBlock_%v_%v", i, j),
				// including
				lookupTables,
				// included
				[]ifaces.Column{
					eho.PosBlock[j], eho.Blocks[i][j],
				},
				// no filters
				nil, nil,
			)

			ctx.comp.GenericFragmentedConditionalInclusion(
				round,
				ctx.queryIDf("HashCheckImportNewState_%v_%v", i, j),
				// including
				lookupTables,
				// included
				[]ifaces.Column{
					eho.PosNewState[j], eho.NewStates[i][j],
				},
				// no filters
				nil, nil,
			)
		}

		ctx.comp.InsertPoseidon2(
			round,
			ctx.queryIDf("HashCheckPoseidon_%v", i),
			eho.Blocks[i],
			eho.OldStates[i],
			eho.NewStates[i],
			selector,
		)
	}
}

// assignHashColumns assigns the hash c olumns.
func (ctx *GenericPlonkProverAction) assignHashColumns(run *wizard.ProverRuntime) {

	var (
		eho         = &ctx.ExternalHasherOption
		posOs       = [poseidon2_koalabear.BlockSize][]field.Element{}
		posBl       = [poseidon2_koalabear.BlockSize][]field.Element{}
		posNs       = [poseidon2_koalabear.BlockSize][]field.Element{}
		sizeHashing = len(posOs)
	)

	for j := 0; j < poseidon2_koalabear.BlockSize; j++ {
		posOs[j] = eho.PosOldState[j].GetColAssignment(run).IntoRegVecSaveAlloc()
		posBl[j] = eho.PosBlock[j].GetColAssignment(run).IntoRegVecSaveAlloc()
		posNs[j] = eho.PosNewState[j].GetColAssignment(run).IntoRegVecSaveAlloc()
	}

	for i := 0; i < ctx.MaxNbInstances; i++ {

		var (
			src = []smartvectors.SmartVector{
				ctx.Columns.L[i].GetColAssignment(run),
				ctx.Columns.R[i].GetColAssignment(run),
				ctx.Columns.O[i].GetColAssignment(run),
			}
			sizeLRO  = ctx.Columns.L[i].Size()
			oldState = [poseidon2_koalabear.BlockSize][]field.Element{}
			block    = [poseidon2_koalabear.BlockSize][]field.Element{}
			newState = [poseidon2_koalabear.BlockSize][]field.Element{}
		)

		for j := 0; j < poseidon2_koalabear.BlockSize; j++ {
			oldState[j] = make([]field.Element, sizeHashing)
			block[j] = make([]field.Element, sizeHashing)
			newState[j] = make([]field.Element, sizeHashing)
		}

		for k := 0; k < poseidon2_koalabear.BlockSize; k++ {
			for j := range oldState {

				var (
					osID = int(posOs[k][j].Uint64())
					blID = int(posBl[k][j].Uint64())
					nsID = int(posNs[k][j].Uint64())
					os   = src[osID/sizeLRO].Get(osID % sizeLRO)
					bl   = src[blID/sizeLRO].Get(blID % sizeLRO)
					ns   = src[nsID/sizeLRO].Get(nsID % sizeLRO)
				)

				oldState[k][j] = os
				block[k][j] = bl
				newState[k][j] = ns
			}
		}
		for j := 0; j < poseidon2_koalabear.BlockSize; j++ {
			run.AssignColumn(eho.OldStates[i][j].GetColID(), smartvectors.NewRegular(oldState[j]))
			run.AssignColumn(eho.Blocks[i][j].GetColID(), smartvectors.NewRegular(block[j]))
			run.AssignColumn(eho.NewStates[i][j].GetColID(), smartvectors.NewRegular(newState[j]))
		}
	}
}

// getHashCheckedPositionSV returns the smartvectors containing the position
// of the hash claims in the LRO columns.
func (ctx *CompilationCtx) getHashCheckedPositionSV() (posOS, posBl, posNS smartvectors.SmartVector) {

	var (
		sls         = ctx.Plonk.hashedGetter()
		size        = utils.NextPowerOfTwo(len(sls))
		numRowPlonk = ctx.DomainSize()
	)

	if ctx.ExternalHasherOption.FixedNbRows > 0 {
		fixedNbRow := ctx.ExternalHasherOption.FixedNbRows
		if fixedNbRow < size {
			utils.Panic("the fixed number of rows %v is smaller than the number of hash claims %v", fixedNbRow, len(sls))
		}
		size = ctx.ExternalHasherOption.FixedNbRows
	}

	var (
		ost = make([]field.Element, size)
		blk = make([]field.Element, size)
		nst = make([]field.Element, size)
	)

	for i, ss := range sls {
		ost[i].SetUint64(uint64(ss[0][0] + ss[0][1]*numRowPlonk))
		blk[i].SetUint64(uint64(ss[1][0] + ss[1][1]*numRowPlonk))
		nst[i].SetUint64(uint64(ss[2][0] + ss[2][1]*numRowPlonk))
	}

	for i := len(sls); i < size; i++ {
		ost[i] = ost[i-1]
		blk[i] = blk[i-1]
		nst[i] = nst[i-1]
	}

	return smartvectors.NewRegular(ost), smartvectors.NewRegular(blk), smartvectors.NewRegular(nst)
}
