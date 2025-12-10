package plonkinternal

import (
	"fmt"

	"github.com/consensys/linea-monorepo/prover/crypto/poseidon2_koalabear"
	"github.com/consensys/linea-monorepo/prover/maths/common/smartvectors"
	"github.com/consensys/linea-monorepo/prover/maths/common/vector"
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

	eho.PosOldState = ctx.comp.InsertPrecomputed(ctx.colIDf("HashCheckPositionOS"), posOsSv)
	eho.PosBlock = ctx.comp.InsertPrecomputed(ctx.colIDf("HashCheckPositionBL"), posBlSv)
	eho.PosNewState = ctx.comp.InsertPrecomputed(ctx.colIDf("HashCheckPositionNS"), posNsSv)
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
			eho.OldStates[i][j] = ctx.comp.InsertCommit(round, ctx.colIDf("HashCheckOldState_%v_%v", i, j), size/poseidon2_koalabear.BlockSize, true)
			eho.Blocks[i][j] = ctx.comp.InsertCommit(round, ctx.colIDf("HashCheckBlock_%v_%v", i, j), size/poseidon2_koalabear.BlockSize, true)
			eho.NewStates[i][j] = ctx.comp.InsertCommit(round, ctx.colIDf("HashCheckNewState_%v_%v", i, j), size/poseidon2_koalabear.BlockSize, true)

			// Those are lookups checking that the LRO columns are consistent with
			// the hash claims.

			ctx.comp.GenericFragmentedConditionalInclusion(
				round,
				ctx.queryIDf("HashCheckImportOldState_%v_%v", i, j),
				// including
				lookupTables,
				// included
				[]ifaces.Column{
					eho.PosOldState, eho.OldStates[i][j],
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
					eho.PosBlock, eho.Blocks[i][j],
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
					eho.PosNewState, eho.NewStates[i][j],
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
		posOs       = eho.PosOldState.GetColAssignment(run).IntoRegVecSaveAlloc()
		posBl       = eho.PosBlock.GetColAssignment(run).IntoRegVecSaveAlloc()
		posNs       = eho.PosNewState.GetColAssignment(run).IntoRegVecSaveAlloc()
		sizeHashing = len(posOs)
	)

	fmt.Printf("Assigning hash checked columns, sizeHashing=%v\n", sizeHashing)
	for i := 0; i < ctx.MaxNbInstances; i++ {

		var (
			src = []smartvectors.SmartVector{
				ctx.Columns.L[i].GetColAssignment(run),
				ctx.Columns.R[i].GetColAssignment(run),
				ctx.Columns.O[i].GetColAssignment(run),
			}
			sizeLRO  = ctx.Columns.L[i].Size()
			oldState = make([]field.Element, sizeHashing)
			block    = make([]field.Element, sizeHashing)
			newState = make([]field.Element, sizeHashing)
		)

		for j := range oldState {

			var (
				osID = int(posOs[j].Uint64())
				blID = int(posBl[j].Uint64())
				nsID = int(posNs[j].Uint64())
				os   = src[osID/sizeLRO].Get(osID % sizeLRO)
				bl   = src[blID/sizeLRO].Get(blID % sizeLRO)
				ns   = src[nsID/sizeLRO].Get(nsID % sizeLRO)
			)
			oldState[j] = os
			block[j] = bl
			newState[j] = ns
		}

		var vecOldState, vecBlock, vecNewState [poseidon2_koalabear.BlockSize][]field.Element

		for j := 0; j < poseidon2_koalabear.BlockSize; j++ {
			vecOldState[j] = make([]field.Element, sizeHashing/poseidon2_koalabear.BlockSize)
			vecBlock[j] = make([]field.Element, sizeHashing/poseidon2_koalabear.BlockSize)
			vecNewState[j] = make([]field.Element, sizeHashing/poseidon2_koalabear.BlockSize)
		}
		for k := range oldState {
			vecOldState[k%poseidon2_koalabear.BlockSize][k/poseidon2_koalabear.BlockSize] = oldState[k]
			vecBlock[k%poseidon2_koalabear.BlockSize][k/poseidon2_koalabear.BlockSize] = block[k]
			vecNewState[k%poseidon2_koalabear.BlockSize][k/poseidon2_koalabear.BlockSize] = newState[k]
		}

		for j := 0; j < poseidon2_koalabear.BlockSize; j++ {

			fmt.Printf("Assigned Hash Old State part %v: %v\n", j, vecOldState[j])
			run.AssignColumn(eho.OldStates[i][j].GetColID(), smartvectors.NewRegular(vecOldState[j]))
			run.AssignColumn(eho.Blocks[i][j].GetColID(), smartvectors.NewRegular(vecBlock[j]))
			run.AssignColumn(eho.NewStates[i][j].GetColID(), smartvectors.NewRegular(vecNewState[j]))
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
	fmt.Printf("Hash claims: %v\n", sls)

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

	fmt.Printf("Hash check positions OS: %v\n", vector.Prettify(ost))
	return smartvectors.NewRegular(ost), smartvectors.NewRegular(blk), smartvectors.NewRegular(nst)
}
